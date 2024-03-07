package logic

import (
	"context"
	"database/sql"
	"fmt"
	"go-gqlgen/constants"
	"go-gqlgen/database"
	"go-gqlgen/graph/model"
	"log"

	"math/rand"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"
)

// function to import the data from exel
func ImportDataFromExcel() (*model.ReportCreated, error) {

	//Connect to MySQL database
	db, err := database.ConnectMySQLDB()
	if err != nil {
		log.Fatal(constants.FailedToConnectDb, err)
		return nil, err
	}

	// Connect to Redis
	rdb := database.ConnectRedis()

	// Open the Excel file
	excelFileName := "sampleReport.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		log.Fatalf(constants.ExelError, err)
	}

	// Iterate through sheets
	for _, sheet := range xlFile.Sheets {
		fmt.Printf("Sheet Name: %s\n", sheet.Name)

		// Iterate through rows
		for _, row := range sheet.Rows {

			if len(row.Cells) < 14 {
				fmt.Println("Row doesn't have enough cells")
				continue
			}
			rowData := make([]string, 0)
			// Iterate through cells
			for _, cell := range row.Cells {
				rowData = append(rowData, cell.String())
			}
			fmt.Println(rowData)

			//Inset data into Database
			id, err := InsetDataToDatabase(db, rowData)
			if err != nil {
				log.Printf(constants.FailedToInsertData, err)
				continue
			}

			// Insert data into Redis
			err = InsertDataToRedis(rdb, id, rowData)
			if err != nil {
				log.Printf(constants.FailedToInsertRedisData, err)
			}
		}
		fmt.Println() // Newline after each sheet
	}
	response := &model.ReportCreated{
		Message: constants.DataImportedFromExel,
	}

	return response, nil
}

// insert the exel data to databse
func InsetDataToDatabase(db *sql.DB, rowData []string) (string, error) {

	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%08d", rand.Intn(100000000))

	// Prepare the SQL statement
	query := "INSERT INTO report.injury_report (date, id, injury_location, gender, age_group, incident_type, days_lost, plant, report_type, shift, department, incident_cost, wkday, month, year, is_active, is_deleted) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? ,?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return "", err
	}

	date := rowData[0]
	injuryLocation := rowData[1]
	gender := rowData[2]
	ageGroup := rowData[3]
	incidentType := rowData[4]
	daysLost := rowData[5]
	plant := rowData[6]
	reportType := rowData[7]
	shift := rowData[8]
	department := rowData[9]
	incidentCost := rowData[10]
	wkday := rowData[11]
	month := rowData[12]
	year := rowData[13]

	// Execute the SQL statement with the provided rowData
	_, err = stmt.Exec(date, id, injuryLocation, gender, ageGroup, incidentType, daysLost, plant, reportType, shift, department, incidentCost, wkday, month, year, true, false)
	if err != nil {
		return "", err
	}

	fmt.Println(constants.DataInserted)
	return id, nil
}

// InsertDataToRedis inserts data into Redis
func InsertDataToRedis(rdb *redis.Client, id string, rowData []string) error {
	// Construct Redis key
	key := fmt.Sprintf(id)

	// Convert rowData to map[string]interface{} for Redis hash
	data := make(map[string]interface{})
	data["date"] = rowData[0]
	data["injury_location"] = rowData[1]
	data["gender"] = rowData[2]
	data["age_group"] = rowData[3]
	data["incident_type"] = rowData[4]
	data["days_lost"] = rowData[5]
	data["plant"] = rowData[6]
	data["report_type"] = rowData[7]
	data["shift"] = rowData[8]
	data["department"] = rowData[9]
	data["incident_cost"] = rowData[10]
	data["wkday"] = rowData[11]
	data["month"] = rowData[12]
	data["year"] = rowData[13]

	// Set hash data in Redis
	err := rdb.HMSet(key, data).Err()
	if err != nil {
		return err
	}

	fmt.Println("Data inserted into Redis:", key)
	return nil
}

// function to add the incident report manually
func AddIncidentReport(ctx context.Context, input *model.AddReportInput) ([]*model.IncidentReport, error) {

	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%08d", rand.Intn(100000000))

	// Connect to MySQL database
	db, err := database.ConnectMySQLDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	// Prepare the INSERT query
	query := `
        INSERT INTO report.injury_report (
             date, id, injury_location, gender, age_group, incident_type, days_lost, plant,
            report_type, shift, department, incident_cost, wkday, month, year, is_active, is_deleted
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	// Execute the INSERT query
	_, err = db.ExecContext(ctx, query,
		input.Date, id, input.InjuryLocation, input.Gender, input.AgeGroup,
		input.IncidentType, input.DaysLost, input.Plant, input.ReportType, input.Shift,
		input.Department, input.IncidentCost, input.Wkday, input.Month, input.Year,
		true, false,
	)
	if err != nil {
		return nil, err
	}

	// Create an IncidentReport object with the inserted data and return it
	insertedRecord := &model.IncidentReport{
		ID:             id,
		Date:           *input.Date,
		InjuryLocation: *input.InjuryLocation,
		Gender:         *input.Gender,
		AgeGroup:       *input.AgeGroup,
		IncidentType:   *input.IncidentType,
		DaysLost:       *input.DaysLost,
		Plant:          *input.Plant,
		ReportType:     *input.ReportType,
		Shift:          *input.Shift,
		Department:     *input.Department,
		IncidentCost:   *input.IncidentCost,
		Wkday:          input.Wkday,
		Month:          *input.Month,
		Year:           *input.Year,
	}

	return []*model.IncidentReport{insertedRecord}, nil
}

// function to get the incident reports
func GetIncidentReport(ctx context.Context) ([]*model.IncidentReport, error) {

	// Define a slice to store the incident reports
	var reports []*model.IncidentReport

	//Connect to MySQL database
	db, err := database.ConnectMySQLDB()
	if err != nil {
		return nil, fmt.Errorf(constants.FailedToConnectDb, err)
	}

	// Query to retrieve all incident reports
	rows, err := db.QueryContext(ctx, "SELECT * FROM report.injury_report")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the rows and scan each row into an IncidentReport object
	for rows.Next() {
		var report model.IncidentReport
		err := rows.Scan(
			&report.Date,
			&report.ID,
			&report.InjuryLocation,
			&report.Gender,
			&report.AgeGroup,
			&report.IncidentType,
			&report.DaysLost,
			&report.Plant,
			&report.ReportType,
			&report.Shift,
			&report.Department,
			&report.IncidentCost,
			&report.Wkday,
			&report.Month,
			&report.Year,
			&report.IsActive,
			&report.IsDeleted,
		)
		if err != nil {
			log.Println(constants.ErrorWhileScaningRow, err)
			continue
		}
		// Append the scanned incident report to the slice
		reports = append(reports, &report)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

// function to get incident report by id
func GetIncidentReportByID(ctx context.Context, id string) (*model.IncidentReport, error) {

	// Initialize an IncidentReport object to store the result
	var report model.IncidentReport

	// Connect to MySQL database
	db, err := database.ConnectMySQLDB()
	if err != nil {
		return nil, fmt.Errorf(constants.FailedToConnectDb, err)
	}

	// Query to retrieve the incident report by ID
	query := "SELECT * FROM report.injury_report WHERE id = ?"
	row := db.QueryRowContext(ctx, query, id)

	// Scan the row into the IncidentReport object
	err = row.Scan(
		&report.Date,
		&report.ID,
		&report.InjuryLocation,
		&report.Gender,
		&report.AgeGroup,
		&report.IncidentType,
		&report.DaysLost,
		&report.Plant,
		&report.ReportType,
		&report.Shift,
		&report.Department,
		&report.IncidentCost,
		&report.Wkday,
		&report.Month,
		&report.Year,
		&report.IsActive,
		&report.IsDeleted,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf(constants.DataNotFound, id)
		}
		return nil, err
	}
	return &report, nil
}

// function to update report in database and in redis
func UpdateIncidentReport(ctx context.Context, input *model.AddReportInput) (*model.ReportCreated, error) {
	//connect with db
	db, err := database.ConnectMySQLDB()
	if err != nil {
		return nil, fmt.Errorf(constants.FailedToConnectDb, err)
	}

	//check if data exist
	_, err = GetIncidentReportByID(ctx, *input.ID)

	if err != nil {
		response := &model.ReportCreated{
			Message: constants.DataNotFound,
		}
		return response, nil
	}

	// Prepare the UPDATE query
	query := `
        UPDATE report.injury_report 
        SET 
		injury_location = COALESCE(?, injury_location),
            gender = COALESCE(?, gender),
            age_group = COALESCE(?, age_group),
            incident_type = COALESCE(?, incident_type),
            days_lost = COALESCE(?, days_lost),
            plant = COALESCE(?, plant),
            report_type = COALESCE(?, report_type),
            shift = COALESCE(?, shift),
            department = COALESCE(?, department),
            incident_cost = COALESCE(?, incident_cost),
            wkday = COALESCE(?, wkday),
            month = COALESCE(?, month),
            year = COALESCE(?, year)
        WHERE id = ?
    `

	// Execute the UPDATE query
	_, err = db.ExecContext(ctx, query,
		input.InjuryLocation, input.Gender, input.AgeGroup, input.IncidentType,
		input.DaysLost, input.Plant, input.ReportType, input.Shift, input.Department,
		input.IncidentCost, input.Wkday, input.Month, input.Year, input.ID,
	)
	if err != nil {
		return nil, err
	}

	// Construct the response
	response := &model.ReportCreated{
		Message: constants.IncidentReportUpdated,
	}
	return response, nil
}

// function to delete report
func DeleteIncidentReport(ctx context.Context, id string) (*model.ReportCreated, error) {
	// Connect to MySQL database
	db, err := database.ConnectMySQLDB()
	if err != nil {
		return nil, fmt.Errorf(constants.FailedToConnectDb, err)
	}
	//defer db.Close()

	//check if data exist
	_, err = GetIncidentReportByID(ctx, id)

	if err != nil {
		response := &model.ReportCreated{
			Message: constants.DataNotFound,
		}
		return response, nil
	}

	// Query to delete the incident report
	query := `DELETE FROM report.injury_report WHERE id = ?`
	_, err = db.ExecContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	// Construct the response
	response := &model.ReportCreated{
		Message: constants.IncidentReportDeleted,
	}
	return response, err
}
