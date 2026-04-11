package constants

// Gmail Sync Constants
const (
	// CDSL India email configuration
	CDSLSenderEmail  = "services@cdslindia.co.in"
	CDSLEmailSubject = "Transactions In Your Demat Account"

	// Gmail API scopes
	GmailScopeReadonly = "https://www.googleapis.com/auth/gmail.readonly"

	// Email parsing patterns
	MaxEmailsToFetch = 100 // Maximum number of emails to fetch in one sync

	// Transaction types
	TransactionTypeBuy  = "BUY"
	TransactionTypeSell = "SELL"

	// Investment categories from email
	CategoryStock      = "EQUITY"
	CategoryMutualFund = "MUTUAL FUND"
)

// Email parsing field names
const (
	FieldTicker    = "ticker"
	FieldQuantity  = "quantity"
	FieldPrice     = "price"
	FieldDate      = "date"
	FieldTransType = "transactionType"
	FieldFundName  = "fundName"
	FieldUnits     = "units"
	FieldNAV       = "nav"
	FieldAmount    = "amount"
	FieldFees      = "fees"
)
