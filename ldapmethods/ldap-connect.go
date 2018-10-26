package ldapmethods

import (
	"fmt"
	"strconv"
	"strings"

	logrus "github.com/sirupsen/logrus"
	ldap "gopkg.in/ldap.v2"
)

// Fields is a slice of maps where the key is the device var ID and key is the value

// ConnectionDetails For LDAP
type ConnectionDetails struct {
	CustomerID  string `json:"customer_id"`
	Host        string
	Port        string
	CN          string
	BaseDN      string `json:"base_dn"`
	Identifier  string
	Password    string
	RequestID   string                 `json:"request_id"`
	Fields      []map[string]string    `json:"fields,omitempty"`
	QueryParams map[string]interface{} `json:"query_params,omitempty"`
}

// LDAPConnectionBind Returns LDAP Connection Binding
func LDAPConnectionBind(connectionDetails *ConnectionDetails) (*ldap.Conn, error) {
	// By default, Port 389 should support SSL. If not, the user can define port 636 which is SSL.
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%v", connectionDetails.Host, connectionDetails.Port))
	if err != nil {
		return nil, err
	}

	// Create LDAP Binding
	err = conn.Bind(connectionDetails.Identifier, connectionDetails.Password)
	if err != nil {
		return nil, err
	}

	// Return connection binding
	return conn, nil
}

// Convert String to a unsigned 32 bit integer
func convertStringToUint32(stringInt string) (uint32, error) {
	uInt64, err := strconv.ParseUint(stringInt, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(uInt64), nil
}

// Build concatinated byte slice of all filter options
// (&(attribute=value/regex)(attribute=value))
func buildSearchTermsString(connectionDetails *ConnectionDetails) string {
	var searchQuery strings.Builder
	for serachAttribute, searchTerm := range connectionDetails.QueryParams {
		searchQuery.WriteString(fmt.Sprintf("(%v=%v)", serachAttribute, searchTerm))
	}
	return searchQuery.String()
}

// GetEntries Return results from LDAP
func GetEntries(connectionDetails *ConnectionDetails) ([]map[string]interface{}, error) {

	conn, err := LDAPConnectionBind(connectionDetails)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	searchQuery := buildSearchTermsString(connectionDetails)
	// pageSizeuint, err := convertStringToUint32(connectionDetails.Limit)
	// if err != nil {
	// 	return nil, err
	// }
	pagingControl := ldap.NewControlPaging(1000)

	var recordTotal int
	for {
		// Make Search Request defining base DN, attributes and filters
		searchRequest := ldap.NewSearchRequest(
			fmt.Sprintf("cn=%v,dc=%v,dc=com,dc=local", connectionDetails.CN, connectionDetails.BaseDN),
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%v)", searchQuery),
			[]string{},
			[]ldap.Control{pagingControl},
		)

		records, err := conn.Search(searchRequest)
		if err != nil {
			return nil, err
		}

		logrus.WithFields(logrus.Fields{
			"LDAP AD Batch Count": len(records.Entries),
		}).Info("LDAP Contacts batch count")

		recordTotal += len(records.Entries)

		// TODO: Add logic for saving and such
		// TODO: Pass

		// Loop through fields and map and store.
		updatedControl := ldap.FindControl(records.Controls, ldap.ControlTypePaging)
		if ctrl, ok := updatedControl.(*ldap.ControlPaging); ctrl != nil && ok && len(ctrl.Cookie) != 0 {
			pagingControl.SetCookie(ctrl.Cookie)
			continue
		}

		break

	}

	fmt.Println(recordTotal)
	// // Make Search Request
	// sr, err := conn.Search(searchRequest)
	// if err != nil {
	// 	return nil, err
	// }

	//TODO: We don't need this, we just need to return a status
	// // Create list of maps
	// entryList := make([]map[string]interface{}, 0)
	// // Iterate through AD Records
	// for i, entry := range sr.Entries {

	// 	// Create new map item and append to list
	// 	entryList = append(entryList, make(map[string]interface{}))

	// 	// Iterate through requestedFields
	// 	for _, field := range connectionDetails.Fields {

	// 		fieldValue := entry.GetAttributeValue(field)

	// 		uuid := uuid.CreateUUID()
	// 		entryList[i]["uuid"] = uuid
	// 		// Check for empty fields and assign nil if empty
	// 		if fieldValue != "" {
	// 			entryList[i][field] = fieldValue
	// 		} else {
	// 			entryList[i][field] = nil
	// 		}
	// 	}
	// }
	m := make([]map[string]interface{}, 0)
	return m, nil
}

// GetEntryAttributes Returns attribute field lists for an entry
func GetEntryAttributes(connectionDetails *ConnectionDetails) ([]string, error) {

	conn, err := LDAPConnectionBind(connectionDetails)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // Defer until end of function

	// Make Search request
	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("dc=%v,dc=com,dc=local", connectionDetails.BaseDN),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&)", //TODO We don't need to filter?
		[]string{},
		nil,
	)

	// Make Search Request
	sr, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	// Assign Attributes slice to var
	attributesSlice := sr.Entries[0].Attributes

	// Create New Slice of attribute names and return
	var attributeNames []string
	for _, attribute := range attributesSlice {
		attributeNames = append(attributeNames, attribute.Name)
	}
	return attributeNames, nil
}
