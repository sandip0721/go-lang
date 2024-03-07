package constants

// Common Success Message

const (
	dataFetch             = "Data fetch Successfully!"
	IncidentReportUpdated = "Incident report updated successfully!"
	IncidentReportDeleted = "Incident report deleted successfully!"
	DataImportedFromExel  = "Data successfully imported from exel"
	DataInserted          = "Data inserted successfully."
	DatabaseConnected     = "Database connected successfully!"
)

// Common Error Message

const (
	failedToFetchData         = "Failed to fetch data"
	FailedToConnectDb         = "Failed to connect to the database: %s"
	FaieldToGetIncidentReport = "failed to get Incident Report: %s"
	FailedToInsertData        = "Error inserting data to database: %v"
	ExelError                 = "Error opening Excel file: %v"
	ErrorWhileScaningRow      = "Error scanning row:"
	DataNotFound              = "data with ID %s not found"
	FailedToInsertRedisData   = "Failed to insert data into redis: %s"
)
