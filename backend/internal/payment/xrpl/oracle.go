package xrpl

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PriceOracle aggregates XRP/USD exchange rates from multiple sources.
type PriceOracle struct {
	sources          []ExchangeSource
	cache            *ExchangeRate
	cacheMu          sync.RWMutex
	logger           *zap.Logger
	updateInterval   time.Duration
	cacheDuration    time.Duration
	outlierThreshold float64
	minSources       int
	stopChan         chan struct{}
	httpClient       *http.Client
}

// OracleConfig holds price oracle configuration.
type OracleConfig struct {
	UpdateInterval   time.Duration
	CacheDuration    time.Duration
	OutlierThreshold float64 // e.g., 0.10 for 10%
	MinSources       int
	HTTPTimeout      time.Duration
}

// DefaultOracleConfig returns default oracle configuration.
func DefaultOracleConfig() *OracleConfig {
	return &OracleConfig{
		UpdateInterval:   30 * time.Second,
		CacheDuration:    60 * time.Second,
		OutlierThreshold: 0.10, // 10%
		MinSources:       2,
		HTTPTimeout:      10 * time.Second,
	}
}

// NewPriceOracle creates a new price oracle.
func NewPriceOracle(cfg *OracleConfig, logger *zap.Logger) *PriceOracle {
	if cfg == nil {
		cfg = DefaultOracleConfig()
	}

	httpClient := &http.Client{
		Timeout: cfg.HTTPTimeout,
	}

	oracle := &PriceOracle{
		sources:          make([]ExchangeSource, 0),
		updateInterval:   cfg.UpdateInterval,
		cacheDuration:    cfg.CacheDuration,
		outlierThreshold: cfg.OutlierThreshold,
		minSources:       cfg.MinSources,
		stopChan:         make(chan struct{}),
		httpClient:       httpClient,
		logger:           logger.With(zap.String("component", "price-oracle")),
	}

	// Register exchange sources
	oracle.sources = append(oracle.sources,
		NewCoinbaseSource(httpClient),
		NewBinanceSource(httpClient),
		NewKrakenSource(httpClient),
		NewBitstampSource(httpClient),
	)

	return oracle
}

// Start begins the automatic price update cycle.
func (o *PriceOracle) Start(ctx context.Context) {
	o.logger.Info("starting price oracle",
		zap.Duration("update_interval", o.updateInterval),
		zap.Int("sources", len(o.sources)),
	)

	// Initial fetch
	if err := o.RefreshRates(ctx); err != nil {
		o.logger.Error("initial rate fetch failed", zap.Error(err))
	}

	// Start background refresh
	go o.backgroundRefresh(ctx)
}

// Stop stops the automatic price update cycle.
func (o *PriceOracle) Stop() {
	close(o.stopChan)
	o.logger.Info("price oracle stopped")
}

// backgroundRefresh periodically updates exchange rates.
func (o *PriceOracle) backgroundRefresh(ctx context.Context) {
	ticker := time.NewTicker(o.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := o.RefreshRates(ctx); err != nil {
				o.logger.Error("rate refresh failed", zap.Error(err))
			}
		case <-o.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

// GetExchangeRate returns the cached exchange rate.
func (o *PriceOracle) GetExchangeRate(ctx context.Context) (*ExchangeRate, error) {
	o.cacheMu.RLock()
	cache := o.cache
	o.cacheMu.RUnlock()

	if cache == nil {
		return nil, fmt.Errorf("exchange rate not initialized")
	}

	// Check if cache is stale
	if time.Since(cache.Updated) > o.cacheDuration {
		o.logger.Warn("cached rate is stale, refreshing",
			zap.Duration("age", time.Since(cache.Updated)),
		)
		if err := o.RefreshRates(ctx); err != nil {
			o.logger.Error("rate refresh failed, using stale cache", zap.Error(err))
			// Return stale cache as fallback
		}
	}

	return o.cache, nil
}

// RefreshRates fetches rates from all sources and updates cache.
func (o *PriceOracle) RefreshRates(ctx context.Context) error {
	o.logger.Debug("fetching exchange rates from sources")

	// Fetch from all sources concurrently
	type result struct {
		source ExchangeSource
		rate   *SourceRate
		err    error
	}

	results := make(chan result, len(o.sources))
	var wg sync.WaitGroup

	for _, source := range o.sources {
		wg.Add(1)
		go func(s ExchangeSource) {
			defer wg.Done()
			rate, err := s.GetRate(ctx)
			results <- result{source: s, rate: rate, err: err}
		}(source)
	}

	// Wait for all sources to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect successful results
	var rates []float64
	var sourceRates []SourceRate
	sourceStatus := make(map[string]string)

	for r := range results {
		if r.err != nil {
			o.logger.Warn("failed to get rate from source",
				zap.String("source", r.source.Name()),
				zap.Error(r.err),
			)
			sourceStatus[r.source.Name()] = fmt.Sprintf("failed: %v", r.err)
			continue
		}

		if r.rate != nil {
			rates = append(rates, r.rate.Rate)
			sourceRates = append(sourceRates, *r.rate)
			sourceStatus[r.source.Name()] = "success"

			o.logger.Debug("received rate from source",
				zap.String("source", r.source.Name()),
				zap.Float64("rate", r.rate.Rate),
			)
		}
	}

	// Check minimum sources requirement
	if len(rates) < o.minSources {
		return fmt.Errorf("insufficient sources: got %d, need at least %d", len(rates), o.minSources)
	}

	// Calculate median
	median := o.calculateMedian(rates)

	// Remove outliers
	filteredRates, filteredSources := o.removeOutliers(rates, sourceRates, median)

	if len(filteredRates) < o.minSources {
		o.logger.Warn("too many outliers, using all rates",
			zap.Int("filtered", len(filteredRates)),
			zap.Int("original", len(rates)),
		)
		filteredRates = rates
		filteredSources = sourceRates
	}

	// Calculate final average
	finalRate := o.calculateAverage(filteredRates)

	// Update cache
	exchangeRate := &ExchangeRate{
		XRPUSD:     finalRate,
		USDXRP:     1.0 / finalRate,
		Sources:    filteredSources,
		MedianRate: median,
		Updated:    time.Now(),
	}

	o.cacheMu.Lock()
	o.cache = exchangeRate
	o.cacheMu.Unlock()

	o.logger.Info("exchange rate updated",
		zap.Float64("rate", finalRate),
		zap.Float64("median", median),
		zap.Int("sources", len(filteredSources)),
	)

	return nil
}

// ConvertUSDtoXRP converts USD amount to XRP with optional volatility buffer.
func (o *PriceOracle) ConvertUSDtoXRP(ctx context.Context, usdAmount float64, includeBuffer bool) (float64, error) {
	rate, err := o.GetExchangeRate(ctx)
	if err != nil {
		return 0, err
	}

	xrpAmount := usdAmount / rate.XRPUSD

	if includeBuffer {
		// Add 2% volatility buffer
		xrpAmount *= 1.02
	}

	// Round to 6 decimal places
	return math.Round(xrpAmount*1000000) / 1000000, nil
}

// ConvertXRPtoUSD converts XRP amount to USD.
func (o *PriceOracle) ConvertXRPtoUSD(ctx context.Context, xrpAmount float64) (float64, error) {
	rate, err := o.GetExchangeRate(ctx)
	if err != nil {
		return 0, err
	}

	usdAmount := xrpAmount * rate.XRPUSD

	// Round to 2 decimal places
	return math.Round(usdAmount*100) / 100, nil
}

// calculateMedian calculates the median of a slice of floats.
func (o *PriceOracle) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// calculateAverage calculates the average of a slice of floats.
func (o *PriceOracle) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// removeOutliers filters out rates that deviate too much from median.
func (o *PriceOracle) removeOutliers(rates []float64, sources []SourceRate, median float64) ([]float64, []SourceRate) {
	filtered := make([]float64, 0)
	filteredSources := make([]SourceRate, 0)

	for i, rate := range rates {
		deviation := math.Abs(rate-median) / median
		if deviation <= o.outlierThreshold {
			filtered = append(filtered, rate)
			filteredSources = append(filteredSources, sources[i])
		} else {
			o.logger.Warn("outlier detected",
				zap.String("source", sources[i].Source),
				zap.Float64("rate", rate),
				zap.Float64("median", median),
				zap.Float64("deviation_pct", deviation*100),
			)
		}
	}

	return filtered, filteredSources
}

// ExchangeSource is an interface for exchange rate sources.
type ExchangeSource interface {
	Name() string
	GetRate(ctx context.Context) (*SourceRate, error)
}

// SourceRate represents a rate from a single source.
type SourceRate struct {
	Source      string
	Rate        float64
	Bid         float64
	Ask         float64
	Volume24h   float64
	LastUpdated time.Time
}

// ExchangeRate represents aggregated exchange rate data.
type ExchangeRate struct {
	XRPUSD     float64
	USDXRP     float64
	Sources    []SourceRate
	MedianRate float64
	Updated    time.Time
}

// CoinbaseSource fetches rates from Coinbase API.
type CoinbaseSource struct {
	client *http.Client
}

func NewCoinbaseSource(client *http.Client) *CoinbaseSource {
	return &CoinbaseSource{client: client}
}

func (s *CoinbaseSource) Name() string {
	return "coinbase"
}

func (s *CoinbaseSource) GetRate(ctx context.Context) (*SourceRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.coinbase.com/v2/prices/XRP-USD/spot", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Data struct {
			Base     string `json:"base"`
			Currency string `json:"currency"`
			Amount   string `json:"amount"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var rate float64
	if _, err := fmt.Sscanf(data.Data.Amount, "%f", &rate); err != nil {
		return nil, fmt.Errorf("failed to parse rate: %w", err)
	}

	return &SourceRate{
		Source:      "coinbase",
		Rate:        rate,
		LastUpdated: time.Now(),
	}, nil
}

// BinanceSource fetches rates from Binance API.
type BinanceSource struct {
	client *http.Client
}

func NewBinanceSource(client *http.Client) *BinanceSource {
	return &BinanceSource{client: client}
}

func (s *BinanceSource) Name() string {
	return "binance"
}

func (s *BinanceSource) GetRate(ctx context.Context) (*SourceRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.binance.com/api/v3/ticker/price?symbol=XRPUSDT", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var rate float64
	if _, err := fmt.Sscanf(data.Price, "%f", &rate); err != nil {
		return nil, fmt.Errorf("failed to parse rate: %w", err)
	}

	return &SourceRate{
		Source:      "binance",
		Rate:        rate,
		LastUpdated: time.Now(),
	}, nil
}

// KrakenSource fetches rates from Kraken API.
type KrakenSource struct {
	client *http.Client
}

func NewKrakenSource(client *http.Client) *KrakenSource {
	return &KrakenSource{client: client}
}

func (s *KrakenSource) Name() string {
	return "kraken"
}

func (s *KrakenSource) GetRate(ctx context.Context) (*SourceRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.kraken.com/0/public/Ticker?pair=XRPUSD", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Result map[string]struct {
			C []string `json:"c"` // Last trade closed
			B []string `json:"b"` // Bid
			A []string `json:"a"` // Ask
			V []string `json:"v"` // Volume
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// Kraken uses XXRPZUSD or XRPUSD
	var ticker interface{}
	for key := range data.Result {
		ticker = data.Result[key]
		break
	}

	if ticker == nil {
		return nil, fmt.Errorf("no ticker data found")
	}

	tickerData := ticker.(struct {
		C []string `json:"c"`
		B []string `json:"b"`
		A []string `json:"a"`
		V []string `json:"v"`
	})

	var rate float64
	if len(tickerData.C) > 0 {
		if _, err := fmt.Sscanf(tickerData.C[0], "%f", &rate); err != nil {
			return nil, fmt.Errorf("failed to parse rate: %w", err)
		}
	}

	return &SourceRate{
		Source:      "kraken",
		Rate:        rate,
		LastUpdated: time.Now(),
	}, nil
}

// BitstampSource fetches rates from Bitstamp API.
type BitstampSource struct {
	client *http.Client
}

func NewBitstampSource(client *http.Client) *BitstampSource {
	return &BitstampSource{client: client}
}

func (s *BitstampSource) Name() string {
	return "bitstamp"
}

func (s *BitstampSource) GetRate(ctx context.Context) (*SourceRate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.bitstamp.net/api/v2/ticker/xrpusd/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data struct {
		Last   string `json:"last"`
		Bid    string `json:"bid"`
		Ask    string `json:"ask"`
		Volume string `json:"volume"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var rate float64
	if _, err := fmt.Sscanf(data.Last, "%f", &rate); err != nil {
		return nil, fmt.Errorf("failed to parse rate: %w", err)
	}

	return &SourceRate{
		Source:      "bitstamp",
		Rate:        rate,
		LastUpdated: time.Now(),
	}, nil
}
