// Tests for the tax service.
//
// This is the layer that turns a manufacturer's production report into an
// amount of money owed to a state. It had no tests. Two of the tests below
// document defects rather than assert correct behaviour; they are named so
// that is obvious, and each says what "correct" would be. Pinning a defect in
// a passing test is not the same as endorsing it — it means the next person to
// change this code finds out immediately, instead of a manufacturer finding
// out from an assessment notice.

package service

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"testing"
	"time"

	"github.com/excise-tax-portal/backend/internal/tax/model"
	"github.com/excise-tax-portal/backend/internal/tax/repository"
	"github.com/excise-tax-portal/backend/pkg/errors"
	"github.com/excise-tax-portal/backend/pkg/logger"
)

// fakeTaxRepo is a hand-written stand-in for TaxRepository.
//
// Only the methods the tests exercise do anything; the rest satisfy the
// interface. rates is keyed "productType/unitType", and a missing key models
// the real failure this service has to cope with — a rate that is not on file
// for the period being reported.
type fakeTaxRepo struct {
	rates map[string]*model.TaxRate
}

func (f *fakeTaxRepo) GetTaxRates(_ context.Context, productType, unitType string, _ time.Time) (*model.TaxRate, error) {
	if rate, ok := f.rates[productType+"/"+unitType]; ok {
		return rate, nil
	}
	return nil, errors.NewNotFoundError("tax_rate", "no rate on file")
}

func (f *fakeTaxRepo) CreateReport(context.Context, *model.TaxReport) error { return nil }
func (f *fakeTaxRepo) GetReportByID(context.Context, int64) (*model.TaxReport, error) {
	return nil, errors.NewNotFoundError("report", "not found")
}
func (f *fakeTaxRepo) GetReportsByManufacturer(context.Context, *model.ReportFilter) ([]*model.TaxReport, error) {
	return nil, nil
}
func (f *fakeTaxRepo) UpdateReport(context.Context, int64, *model.TaxReport) error { return nil }
func (f *fakeTaxRepo) SubmitReport(context.Context, int64) error                   { return nil }
func (f *fakeTaxRepo) GetAllCurrentTaxRates(context.Context) ([]*model.TaxRate, error) {
	return nil, nil
}
func (f *fakeTaxRepo) CheckOverlappingReports(context.Context, int64, time.Time, time.Time, *int64) (bool, error) {
	return false, nil
}
func (f *fakeTaxRepo) CountReportsByFilter(context.Context, *model.ReportFilter) (int64, error) {
	return 0, nil
}
func (f *fakeTaxRepo) CalculateTaxFromProduction(context.Context, json.RawMessage) (float64, error) {
	return 0, nil
}

var _ repository.TaxRepository = (*fakeTaxRepo)(nil)

func newTestService(t *testing.T, rates map[string]*model.TaxRate) TaxService {
	t.Helper()
	log, err := logger.NewLogger(logger.DefaultConfig())
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	return NewTaxService(&fakeTaxRepo{rates: rates}, log)
}

func production(t *testing.T, items ...model.ProductionItem) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(model.ProductionData{Items: items})
	if err != nil {
		t.Fatalf("marshal production data: %v", err)
	}
	return raw
}

// Money comparisons need a tolerance: the service computes in float64, so
// exact equality is the wrong assertion even when the arithmetic is right.
// (That the money path uses float64 at all is noted in the README as a real
// concern; these tests pin current behaviour rather than pretend otherwise.)
func approxEqual(a, b float64) bool {
	const epsilon = 1e-9
	diff := a - b
	return diff < epsilon && diff > -epsilon
}

func TestCalculateTaxAmount_FlatRatePerUnit(t *testing.T) {
	svc := newTestService(t, map[string]*model.TaxRate{
		"beer/barrel": {ProductType: "beer", UnitType: "barrel", RatePerUnit: 3.50},
	})

	got, err := svc.CalculateTaxAmount(context.Background(), production(t, model.ProductionItem{
		ProductType: model.ProductTypeBeer,
		ProductName: "Pale Ale",
		UnitType:    "barrel",
		Quantity:    100,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(got.TotalTaxAmount, 350.0) {
		t.Errorf("total = %v, want 350", got.TotalTaxAmount)
	}
	if len(got.BreakdownByProduct) != 1 {
		t.Fatalf("breakdown has %d entries, want 1", len(got.BreakdownByProduct))
	}
	if got.BreakdownByProduct[0].RatePerUnit != 3.50 {
		t.Errorf("breakdown rate = %v, want 3.50", got.BreakdownByProduct[0].RatePerUnit)
	}
}

func TestCalculateTaxAmount_SpiritsUseProofGallons(t *testing.T) {
	// Spirits are taxed per proof gallon, not per wine gallon. Proof is twice
	// ABV, so 40% ABV is 80 proof and a gallon of it is 0.8 proof gallons.
	// 100 gallons * 0.8 * $13.50 = $1080.
	svc := newTestService(t, map[string]*model.TaxRate{
		"spirits/gallon": {ProductType: "spirits", UnitType: "gallon", RatePerUnit: 13.50},
	})

	got, err := svc.CalculateTaxAmount(context.Background(), production(t, model.ProductionItem{
		ProductType: model.ProductTypeSpirits,
		ProductName: "Bourbon",
		UnitType:    "gallon",
		Quantity:    100,
		ABV:         40,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approxEqual(got.TotalTaxAmount, 1080.0) {
		t.Errorf("total = %v, want 1080", got.TotalTaxAmount)
	}
}

func TestCalculateTaxAmount_SpiritsWithoutABVFallBackToFlatRate(t *testing.T) {
	// ABV is optional in the model. When it is absent the proof adjustment is
	// skipped entirely and the spirit is taxed per wine gallon — which is a
	// LOWER number than the proof-gallon rule produces for anything over 50%
	// ABV. Worth pinning: a missing optional field quietly changing which tax
	// rule applies is the kind of thing that survives a long time unnoticed.
	svc := newTestService(t, map[string]*model.TaxRate{
		"spirits/gallon": {ProductType: "spirits", UnitType: "gallon", RatePerUnit: 13.50},
	})

	got, err := svc.CalculateTaxAmount(context.Background(), production(t, model.ProductionItem{
		ProductType: model.ProductTypeSpirits,
		UnitType:    "gallon",
		Quantity:    100,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !approxEqual(got.TotalTaxAmount, 1350.0) {
		t.Errorf("total = %v, want 1350 (flat rate, no proof adjustment)", got.TotalTaxAmount)
	}
}

func TestCalculateTaxAmount_SumsAcrossProducts(t *testing.T) {
	svc := newTestService(t, map[string]*model.TaxRate{
		"beer/barrel":    {RatePerUnit: 3.50},
		"wine/gallon":    {RatePerUnit: 1.07},
		"spirits/gallon": {RatePerUnit: 13.50},
	})

	got, err := svc.CalculateTaxAmount(context.Background(), production(t,
		model.ProductionItem{ProductType: model.ProductTypeBeer, UnitType: "barrel", Quantity: 10},
		model.ProductionItem{ProductType: model.ProductTypeWine, UnitType: "gallon", Quantity: 100},
		model.ProductionItem{ProductType: model.ProductTypeSpirits, UnitType: "gallon", Quantity: 10, ABV: 50},
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 35 + 107 + (10 * 1.0 * 13.50) = 277
	if !approxEqual(got.TotalTaxAmount, 277.0) {
		t.Errorf("total = %v, want 277", got.TotalTaxAmount)
	}
	if len(got.BreakdownByProduct) != 3 {
		t.Errorf("breakdown has %d entries, want 3", len(got.BreakdownByProduct))
	}
}

func TestCalculateTaxAmount_MissingRateSilentlyUnderreports_KnownDefect(t *testing.T) {
	// DEFECT, pinned deliberately.
	//
	// When no tax rate is on file, CalculateTaxAmount logs a warning and
	// `continue`s. The product vanishes from both the breakdown and the total,
	// and the caller receives a TaxCalculation that looks complete and
	// authoritative but understates what is owed. Nothing in the return value
	// says a product was dropped.
	//
	// Failing toward a smaller tax bill, silently, is the worst available
	// direction for this to be wrong: it produces a filing that is incorrect
	// in the manufacturer's favour, which is the kind of error that becomes a
	// penalty rather than a refund.
	//
	// Correct behaviour would be to return an error, or to surface the
	// unpriced items on the TaxCalculation so the caller must deal with them.
	// This test asserts what the code does today and will fail the moment that
	// changes, which is when someone should read this comment.
	svc := newTestService(t, map[string]*model.TaxRate{
		"beer/barrel": {RatePerUnit: 3.50},
		// no rate for spirits
	})

	got, err := svc.CalculateTaxAmount(context.Background(), production(t,
		model.ProductionItem{ProductType: model.ProductTypeBeer, UnitType: "barrel", Quantity: 100},
		model.ProductionItem{ProductType: model.ProductTypeSpirits, UnitType: "gallon", Quantity: 500, ABV: 40},
	))
	if err != nil {
		t.Fatalf("no error is returned today: %v", err)
	}

	if !approxEqual(got.TotalTaxAmount, 350.0) {
		t.Errorf("total = %v; the unpriced spirits are dropped, leaving only the beer", got.TotalTaxAmount)
	}
	if len(got.BreakdownByProduct) != 1 {
		t.Errorf("breakdown has %d entries; the dropped product leaves no trace", len(got.BreakdownByProduct))
	}
}

func TestCalculateTaxAmount_RejectsMalformedProductionData(t *testing.T) {
	svc := newTestService(t, nil)

	_, err := svc.CalculateTaxAmount(context.Background(), json.RawMessage(`{"items": "not-an-array"}`))
	if err == nil {
		t.Fatal("expected an error for malformed production data")
	}
	// pkg/errors exposes IsNotFound, IsValidation, IsUnauthorized, IsForbidden,
	// IsConflict and IsInternal — but no IsBadRequest, so this asserts on the
	// code directly. Worth adding for symmetry; not done here to keep this
	// change to tests only.
	var appErr *errors.AppError
	if !stderrors.As(err, &appErr) || appErr.Code != errors.ErrCodeBadRequest {
		t.Errorf("error = %v, want a BAD_REQUEST AppError", err)
	}
}

func TestCalculateTaxAmount_EmptyProductionIsZero(t *testing.T) {
	svc := newTestService(t, nil)

	got, err := svc.CalculateTaxAmount(context.Background(), production(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalTaxAmount != 0 {
		t.Errorf("total = %v, want 0", got.TotalTaxAmount)
	}
	// An empty breakdown must be an empty slice, not nil: it is serialised to
	// JSON for the portal, and `null` and `[]` are not the same thing to a
	// client iterating over it.
	if got.BreakdownByProduct == nil {
		t.Error("breakdown is nil; should be an empty slice for JSON callers")
	}
}

func TestCanAccessReport_GrantsEveryoneAccess_KnownDefect(t *testing.T) {
	// DEFECT, pinned deliberately.
	//
	// canAccessReport returns true on every path. The admin/reviewer branch is
	// real; the fallthrough for everyone else is a placeholder whose own
	// comment says "we'll assume manufacturers can access their own reports"
	// — but nothing checks that the caller is associated with the manufacturer
	// on the report.
	//
	// The consequence is that any authenticated user can read any other
	// manufacturer's tax report by id. In a system where the reports are
	// commercially sensitive production volumes belonging to competing
	// businesses, that is a tenant-isolation failure, not a rough edge.
	//
	// Fixing it needs a user-to-manufacturer association that the schema does
	// not currently carry, so it is a real change rather than a one-line
	// patch. This test exists so the gap is impossible to miss.
	svc := newTestService(t, nil).(*taxService)

	report := &model.TaxReport{ID: 1, ManufacturerID: 42}

	if !svc.canAccessReport(report, 999, "manufacturer") {
		t.Error("behaviour changed: unassociated user is now denied — update this test and delete the defect note")
	}
	if !svc.canAccessReport(report, 999, "") {
		t.Error("behaviour changed: an empty role is now denied — update this test and delete the defect note")
	}
}

func TestValidateCreateRequest(t *testing.T) {
	svc := newTestService(t, nil).(*taxService)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		req     *model.CreateReportRequest
		wantErr bool
	}{
		{
			name: "valid monthly report",
			req: &model.CreateReportRequest{
				ManufacturerID:    1,
				ReportType:        "monthly",
				ReportPeriodStart: start,
				ReportPeriodEnd:   start.AddDate(0, 1, 0),
			},
		},
		{
			name: "manufacturer id must be positive",
			req: &model.CreateReportRequest{
				ManufacturerID:    0,
				ReportType:        "monthly",
				ReportPeriodStart: start,
				ReportPeriodEnd:   start.AddDate(0, 1, 0),
			},
			wantErr: true,
		},
		{
			name: "report type must be known",
			req: &model.CreateReportRequest{
				ManufacturerID:    1,
				ReportType:        "fortnightly",
				ReportPeriodStart: start,
				ReportPeriodEnd:   start.AddDate(0, 1, 0),
			},
			wantErr: true,
		},
		{
			name: "period end must not precede period start",
			req: &model.CreateReportRequest{
				ManufacturerID:    1,
				ReportType:        "monthly",
				ReportPeriodStart: start,
				ReportPeriodEnd:   start.AddDate(0, -1, 0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateCreateRequest(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
