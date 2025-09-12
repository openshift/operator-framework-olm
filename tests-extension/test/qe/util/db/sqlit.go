// Package db provides SQLite database utilities for testing OLM operator catalogs.
// It includes functionality to query operator bundles, channels, packages, and related images
// from SQLite databases used by OLM (Operator Lifecycle Manager).
package db

import (
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"strings"

	// Import SQLite driver
	_ "github.com/mattn/go-sqlite3"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// Sqlit is a SQLite helper for executing commands on OLM catalog databases.
// It provides methods to query operator bundles, channels, packages, and related images.
type Sqlit struct {
	// driverName specifies the SQL driver to use (typically "sqlite3")
	driverName string
}

// OperatorBundle represents an operator bundle entry in the catalog database.
// It contains metadata about an operator bundle including its name, bundle path, and version.
type OperatorBundle struct {
	// name is the unique identifier for the operator bundle
	name string
	// bundlepath is the path to the bundle in the catalog
	bundlepath string
	// version is the semantic version of the operator bundle
	version string
}

// Channel represents a channel entry in the catalog database.
// Channels define upgrade paths and relationships between operator versions.
type Channel struct {
	// entry_id is the unique identifier for the channel entry
	entry_id int64
	// channel_name is the name of the channel (e.g., "stable", "alpha")
	channel_name string
	// package_name is the name of the operator package this channel belongs to
	package_name string
	// operatorbundle_name is the name of the operator bundle in this channel
	operatorbundle_name string
	// replaces indicates which bundle this bundle replaces in the upgrade path
	replaces string
	// depth represents the position in the channel upgrade graph
	depth int
}

// Package represents an operator package in the catalog database.
// A package groups related operator bundles and defines the default channel.
type Package struct {
	// name is the unique identifier for the operator package
	name string
	// default_channel is the default channel to use for this package
	default_channel string
}

// Image represents a container image related to an operator bundle.
// This includes both the operator image and any related images it references.
type Image struct {
	// image is the full container image reference (registry/repository:tag)
	image string
	// operatorbundle_name is the name of the operator bundle that uses this image
	operatorbundle_name string
}

// NewSqlit creates a new SQLite instance configured for OLM catalog databases.
// Returns a Sqlit struct initialized with the sqlite3 driver.
func NewSqlit() *Sqlit {
	return &Sqlit{
		driverName: "sqlite3",
	}
}

// QueryDB executes a SQLite query on the specified database file and returns the result rows.
// It first checks if the database file exists, then opens a connection and executes the query.
func (c *Sqlit) QueryDB(dbFilePath string, query string) (*sql.Rows, error) {
	// Check if database file exists
	if _, err := os.Stat(dbFilePath); os.IsNotExist(err) {
		e2e.Logf("file %s do not exist", dbFilePath)
		return nil, err
	}
	// Open database connection
	database, err := sql.Open(c.driverName, dbFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := database.Close(); err != nil {
			e2e.Logf("Failed to close database: %v", err)
		}
	}()
	// Execute query and return rows
	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, err
}

// QueryOperatorBundle retrieves all operator bundles from the catalog database.
// Returns a slice of OperatorBundle structs containing name, bundle path, and version information.
func (c *Sqlit) QueryOperatorBundle(dbFilePath string) ([]OperatorBundle, error) {
	// Query all operator bundles from the database
	rows, err := c.QueryDB(dbFilePath, "SELECT name,bundlepath,version FROM operatorbundle")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			e2e.Logf("Failed to close rows: %v", err)
		}
	}()

	// Parse query results into OperatorBundle structs
	var OperatorBundles []OperatorBundle
	var name string
	var bundlepath string
	var version string
	for rows.Next() {
		if err := rows.Scan(&name, &bundlepath, &version); err != nil {
			e2e.Logf("Failed to scan row: %v", err)
			continue
		}
		OperatorBundles = append(OperatorBundles, OperatorBundle{name: name, bundlepath: bundlepath, version: version})
		e2e.Logf("OperatorBundles: name: %s,bundlepath: %s, version: %s", name, bundlepath, version)
	}
	return OperatorBundles, nil
}

// CheckOperatorBundlePathExist checks if an operator bundle with the specified path exists in the catalog.
// Returns true if a bundle with the given bundle path is found, false otherwise.
func (c *Sqlit) CheckOperatorBundlePathExist(dbFilePath string, bundlepath string) (bool, error) {
	// Get all operator bundles from the database
	OperatorBundles, err := c.QueryOperatorBundle(dbFilePath)
	if err != nil {
		return false, err
	}
	// Search for matching bundle path
	for _, OperatorBundle := range OperatorBundles {
		if strings.Compare(OperatorBundle.bundlepath, bundlepath) == 0 {
			return true, nil
		}
	}
	return false, nil
}

// CheckOperatorBundleNameExist checks if an operator bundle with the specified name exists in the catalog.
// Returns true if a bundle with the given name is found, false otherwise.
func (c *Sqlit) CheckOperatorBundleNameExist(dbFilePath string, bundleName string) (bool, error) {
	// Get all operator bundles from the database
	OperatorBundles, err := c.QueryOperatorBundle(dbFilePath)
	if err != nil {
		return false, err
	}
	// Search for matching bundle name
	for _, OperatorBundle := range OperatorBundles {
		if strings.Compare(OperatorBundle.name, bundleName) == 0 {
			return true, nil
		}
	}
	return false, nil
}

// QueryOperatorChannel retrieves all channel entries from the catalog database.
// Returns a slice of Channel structs containing channel information and upgrade relationships.
func (c *Sqlit) QueryOperatorChannel(dbFilePath string) ([]Channel, error) {
	// Query all channel entries from the database
	rows, err := c.QueryDB(dbFilePath, "select * from channel_entry;")
	var (
		Channels            []Channel
		entry_id            int64
		channel_name        string
		package_name        string
		operatorbundle_name string
		replaces            string
		depth               int
	)
	defer func() {
		if err := rows.Close(); err != nil {
			e2e.Logf("Failed to close rows: %v", err)
		}
	}()
	if err != nil {
		return Channels, err
	}

	// Parse query results into Channel structs
	for rows.Next() {
		if err := rows.Scan(&entry_id, &channel_name, &package_name, &operatorbundle_name, &replaces, &depth); err != nil {
			e2e.Logf("Failed to scan row: %v", err)
			continue
		}
		Channels = append(Channels, Channel{entry_id: entry_id,
			channel_name:        channel_name,
			package_name:        package_name,
			operatorbundle_name: operatorbundle_name,
			replaces:            replaces,
			depth:               depth})
	}
	return Channels, nil
}

// QueryPackge retrieves all packages from the catalog database.
// Returns a slice of Package structs containing package names and default channels.
// Note: Function name has a typo (should be QueryPackage), but kept for API compatibility.
func (c *Sqlit) QueryPackge(dbFilePath string) ([]Package, error) {
	// Query all packages from the database
	rows, err := c.QueryDB(dbFilePath, "select * from package;")
	var (
		Packages        []Package
		name            string
		default_channel string
	)
	defer func() {
		if err := rows.Close(); err != nil {
			e2e.Logf("Failed to close rows: %v", err)
		}
	}()
	if err != nil {
		return Packages, err
	}

	// Parse query results into Package structs
	for rows.Next() {
		if err := rows.Scan(&name, &default_channel); err != nil {
			e2e.Logf("Failed to scan row: %v", err)
			continue
		}
		Packages = append(Packages, Package{name: name,
			default_channel: default_channel})
	}
	return Packages, nil
}

// QueryRelatedImage retrieves all related images from the catalog database.
// Returns a slice of Image structs containing image references and their associated operator bundles.
func (c *Sqlit) QueryRelatedImage(dbFilePath string) ([]Image, error) {
	// Query all related images from the database
	rows, err := c.QueryDB(dbFilePath, "select * from related_image;")
	var (
		relatedImages       []Image
		image               string
		operatorbundle_name string
	)
	defer func() {
		if err := rows.Close(); err != nil {
			e2e.Logf("Failed to close rows: %v", err)
		}
	}()
	if err != nil {
		return relatedImages, err
	}

	// Parse query results into Image structs
	for rows.Next() {
		if err := rows.Scan(&image, &operatorbundle_name); err != nil {
			e2e.Logf("Failed to scan row: %v", err)
			continue
		}
		relatedImages = append(relatedImages, Image{image: image,
			operatorbundle_name: operatorbundle_name})
	}
	return relatedImages, nil
}

// GetOperatorChannelByColumn extracts values from a specific column of all channel entries.
// Uses reflection to dynamically access the specified field from Channel structs.
// Returns a slice of strings containing the values from the specified column.
func (c *Sqlit) GetOperatorChannelByColumn(dbFilePath string, column string) ([]string, error) {
	// Get all channel entries from the database
	channels, err := c.QueryOperatorChannel(dbFilePath)
	if err != nil {
		return nil, err
	}
	// Extract values from the specified column using reflection
	var channelList []string
	for _, channel := range channels {
		// Use reflection to get the field value by name
		value := reflect.Indirect(reflect.ValueOf(&channel)).FieldByName(column)
		channelList = append(channelList, value.String())
	}
	return channelList, nil
}

// Query is a generic method to extract values from a specific column of any supported table.
// Supports tables: operatorbundle, channel_entry, package, related_image.
// Uses reflection to dynamically access the specified field from the appropriate struct type.
func (c *Sqlit) Query(dbFilePath string, table string, column string) ([]string, error) {
	var valueList []string
	switch table {
	case "operatorbundle":
		// Query operator bundles and extract the specified column
		result, err := c.QueryOperatorBundle(dbFilePath)
		if err != nil {
			return nil, err
		}
		for _, channel := range result {
			value := reflect.Indirect(reflect.ValueOf(&channel)).FieldByName(column)
			valueList = append(valueList, value.String())
		}
		return valueList, nil
	case "channel_entry":
		// Query channel entries and extract the specified column
		result, err := c.QueryOperatorChannel(dbFilePath)
		if err != nil {
			return nil, err
		}
		for _, channel := range result {
			value := reflect.Indirect(reflect.ValueOf(&channel)).FieldByName(column)
			valueList = append(valueList, value.String())
		}
		return valueList, nil
	case "package":
		// Query packages and extract the specified column
		result, err := c.QueryPackge(dbFilePath)
		if err != nil {
			return nil, err
		}
		for _, packageIndex := range result {
			value := reflect.Indirect(reflect.ValueOf(&packageIndex)).FieldByName(column)
			valueList = append(valueList, value.String())
		}
		return valueList, nil
	case "related_image":
		// Query related images and extract the specified column
		result, err := c.QueryRelatedImage(dbFilePath)
		if err != nil {
			return nil, err
		}
		for _, imageIndex := range result {
			value := reflect.Indirect(reflect.ValueOf(&imageIndex)).FieldByName(column)
			valueList = append(valueList, value.String())
		}
		return valueList, nil
	default:
		err := fmt.Errorf("do not support to query table %s", table)
		return nil, err
	}
}

// DBHas checks if the database contains all the specified values in the given table/column.
// Returns true if all values in valueList are found in the database column, false otherwise.
func (c *Sqlit) DBHas(dbFilePath string, table string, column string, valueList []string) (bool, error) {
	// Get all values from the specified table/column
	valueListDB, err := c.Query(dbFilePath, table, column)
	if err != nil {
		return false, err
	}
	// Check if database contains all specified values
	return contains(valueListDB, valueList), nil
}

// DBMatch checks if the database column exactly matches the specified list of values.
// Returns true if the database contains exactly the same values (same count and content), false otherwise.
func (c *Sqlit) DBMatch(dbFilePath string, table string, column string, valueList []string) (bool, error) {
	// Get all values from the specified table/column
	valueListDB, err := c.Query(dbFilePath, table, column)
	if err != nil {
		return false, err
	}
	// Check if database values exactly match the specified values
	return match(valueListDB, valueList), nil
}

// contains checks if stringList1 contains all elements from stringList2.
// Returns true if all elements in stringList2 are found in stringList1, false otherwise.
func contains(stringList1 []string, stringList2 []string) bool {
	// Check each element in stringList2
	for _, stringIndex2 := range stringList2 {
		containFlag := false
		// Search for the element in stringList1
		for _, stringIndex1 := range stringList1 {
			if strings.Compare(stringIndex1, stringIndex2) == 0 {
				containFlag = true
				break
			}
		}
		// If any element is not found, return false
		if !containFlag {
			e2e.Logf("[%s] do not contain [%s]", strings.Join(stringList1, ","), strings.Join(stringList2, ","))
			return false
		}
	}
	return true
}

// match checks if two string lists contain exactly the same elements (same length and content).
// Returns true if both lists have the same length and all elements match, false otherwise.
func match(stringList1 []string, stringList2 []string) bool {
	// Check if lists have the same length
	if len(stringList1) != len(stringList2) {
		return false
	}
	// Check if all elements in stringList2 exist in stringList1
	for _, stringIndex2 := range stringList2 {
		containFlag := false
		for _, stringIndex1 := range stringList1 {
			if strings.Compare(stringIndex1, stringIndex2) == 0 {
				containFlag = true
				break
			}
		}
		// If any element is not found, lists don't match
		if !containFlag {
			e2e.Logf("[%s] do not equal to [%s]", strings.Join(stringList1, ","), strings.Join(stringList2, ","))
			return false
		}
	}
	return true
}
