package operation

import (
	"errors"
	"fmt"
	"net"
	"project/reportingApplication/utils"
	"time"

	ldap "gopkg.in/ldap.v2"
)

type LDAP interface {
	Authenticate(username, password string)
}

type Config struct {
	BaseDN string
	ROUser User // the read-only user for initial bind
	Host   string
	Filter string
}

// User forms the LDAP user
type User struct {
	Name string
	Pswd string
}

// LdapClient for Ldap User
type LdapClient struct {
	Config
}

type Login struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// Authenticate implementation for the Client interface
func (c *LdapClient) Authenticate(user Login) error {
	// establish connection
	conn, err := c.connectLDAP()
	if err != nil {
		utils.Log("LDAP connection Failed, err :", err.Error())
		return err
	}
	defer conn.Close()

	// perform initial read-only bind
	if err = conn.Bind(c.ROUser.Name, c.ROUser.Pswd); err != nil {
		utils.Log("LDAP Bind Failed, err :", err.Error())
		return err
	}

	var filterVal string

	if user.Username != "" {
		filterVal = user.Username
	} else {
		filterVal = user.Email
	}

	//filter := "(&(objectClass=person)(memberOf:1.2.840.113556.1.4.1941:=CN=Chat,CN=Users,DC=example,DC=com)(|(sAMAccountName={username})(mail={username})))"

	searchRequest := ldap.NewSearchRequest(c.BaseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, fmt.Sprintf("(%v=%v)", c.Filter, filterVal), []string{}, nil)

	// find the user attempting to login
	results, err := conn.Search(searchRequest)
	if err != nil {
		utils.Log("LDAP Search Failed, err :", err.Error())
		return err
	}

	utils.Log("Total Users :", len(results.Entries))

	if len(results.Entries) < 1 {
		err = errors.New("not found")
		utils.Log("LDAP User not Found")
		return err
	}

	// attempt auth
	return conn.Bind(results.Entries[0].DN, user.Password)
}

// NewLDAPClient : creates a new client with the provided config
func NewLDAPClient(config Config) (*LdapClient, error) {
	config, err := validateConfig(config)
	if err != nil {
		utils.Log("LDAP Configuration invalid, err:", err.Error())
		return nil, err
	}
	lClient := &LdapClient{config}
	return lClient, err
}

// establishes a connection with an ldap host
func (c *LdapClient) connectLDAP() (*ldap.Conn, error) {
	conn, err := net.DialTimeout("tcp", c.Host, time.Second*10)
	if err != nil {
		return nil, err
	}
	lCon := ldap.NewConn(conn, false)
	lCon.Start()
	return lCon, nil
}

// validates that all required fields were provided
// handles default value for Filter
func validateConfig(config Config) (Config, error) {
	if config.BaseDN == "" || config.Host == "" || config.ROUser.Name == "" || config.ROUser.Pswd == "" {
		return Config{}, errors.New("[CONFIG] The config provided could not be validated")
	}
	if config.Filter == "" {
		config.Filter = "sAMAccountName"
	}
	return config, nil
}
