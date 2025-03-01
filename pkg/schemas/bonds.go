package schemas

import "time"

type BondDate time.Time

// Custom unmarshal function for time
func (t *BondDate) UnmarshalJSON(data []byte) error {
	// Define the custom time format
	const layout = `"2006-01-02 15:04:05"`
	str := string(data)

	// Remove the surrounding quotes if present
	str = str[1 : len(str)-1]

	// Parse the string to time.Time
	parsedTime, err := time.Parse(layout, str)
	if err != nil {
		return err
	}

	// Set the value to the time.Time field
	*t = BondDate(parsedTime)
	return nil
}

type BondInterest struct {
	IssueCode    string  `json:"issue_code"`
	Year1Coupon  float64 `json:"year1_coupon"`
	Year1Return  float64 `json:"year1_return"`
	Year2Coupon  float64 `json:"year2_coupon"`
	Year2Return  float64 `json:"year2_return"`
	Year3Coupon  float64 `json:"year3_coupon"`
	Year3Return  float64 `json:"year3_return"`
	Year4Coupon  float64 `json:"year4_coupon"`
	Year4Return  float64 `json:"year4_return"`
	Year5Coupon  float64 `json:"year5_coupon"`
	Year5Return  float64 `json:"year5_return"`
	Year6Coupon  float64 `json:"year6_coupon"`
	Year6Return  float64 `json:"year6_return"`
	Year7Coupon  float64 `json:"year7_coupon"`
	Year7Return  float64 `json:"year7_return"`
	Year8Coupon  float64 `json:"year8_coupon"`
	Year8Return  float64 `json:"year8_return"`
	Year9Coupon  float64 `json:"year9_coupon"`
	Year9Return  float64 `json:"year9_return"`
	Year10Coupon float64 `json:"year10_coupon"`
	Year10Return float64 `json:"year10_return"`
}

type SavingsBonds struct {
	IssueCode                string   `json:"issue_code"`
	ISINCode                 string   `json:"isin_code"`
	AuctionTenor             int      `json:"auction_tenor"`               // how many years to maturity
	IssueSize                float64  `json:"issue_size"`                  // in million of dollars
	AmountApplied            float64  `json:"amount_applied"`              // in million of dollars
	TotalAppliedWithinLimits float64  `json:"total_applied_within_limits"` // in million of dollars. Total amount within individual allotment limits.
	AmountAlloted            float64  `json:"amount_alloted"`              // in million of dollars
	RandomAllotedAmount      float64  `json:"rndm_alloted_amt"`
	RandomAllotedRate        float64  `json:"rndm_alloted_rate"`
	CutoffAmount             float64  `json:"cutoff_amt"`
	FirstInterestDate        BondDate `json:"first_int_date"`
	SBInt1                   BondDate `json:"sb_int_1"`      // Month of first interest payment
	SBInt2                   BondDate `json:"sb_int_2"`      // Month of first interest payment
	PaymentMonth             string   `json:"payment_month"` // Interest payment months e.g. "Mar,Sep"
	IssueDate                BondDate `json:"issue_date"`
	MaturityDate             BondDate `json:"maturity_date"`
	AnnDate                  BondDate `json:"ann_date"`
	LastDayToApply           BondDate `json:"last_day_to_apply"`
	TenderDate               BondDate `json:"tender_date"`
	StartOfRedemption        BondDate `json:"start_of_redemption"`
	EndOfRedemption          BondDate `json:"end_of_redemption"`
}
