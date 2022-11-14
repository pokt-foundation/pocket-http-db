package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gojektech/heimdall/httpclient"
	"github.com/lib/pq"
	postgresdriver "github.com/pokt-foundation/portal-api-go/postgres-driver"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/suite"
)

const (
	baseURL = "http://localhost:8080"
	apiKey  = "test_api_key_6789"

	connectionString = "postgres://postgres:pgpassword@localhost:5432/postgres?sslmode=disable"

	testUserID = "test_id_de26a0db3b6c631c4"
)

var (
	ErrResponseNotOK error = errors.New("Response not OK")

	testClient = httpclient.NewClient(httpclient.WithHTTPTimeout(5*time.Second), httpclient.WithRetryCount(0))

	createdBlockchainID  string = "" // Used to create a blockchain Redirect
	createdApplicationID string = "" // Used to create a LoadBalancer.ApplicationIDs slice
)

type PHDTestSuite struct {
	suite.Suite
	PGDriver *postgresdriver.PostgresDriver
}

func (t *PHDTestSuite) SetupSuite() {
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Printf("Problem with listener, error: %s, event type: %d", err.Error(), ev)
		}
	}
	listener := pq.NewListener(connectionString, 10*time.Second, time.Minute, reportProblem)
	pgDriver, err := postgresdriver.NewPostgresDriverFromConnectionString(connectionString, listener)
	t.NoError(err)

	t.PGDriver = pgDriver
}

func TestE2E_RunSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end to end test")
	}

	suite.Run(t, new(PHDTestSuite))
}

/*
To run the E2E suite use the command `make test_e2e` from the repository root.
The E2E suite also runs on all Pull Requests to the main or staging branches.

The End-to-End test suite uses a Dockerized reproduction of Postgres & PHD,
tests all of PHD's endpoints using an HTTP client and the verifies the results.
*/

func (t *PHDTestSuite) TestPHD_BlockchainEndpoints() {
	/* Create Blockchain -> POST /blockchain */
	createdBlockchain, err := post[repository.Blockchain]("blockchain", []byte(blockchainJSON))
	t.NoError(err)
	t.blockchainAssertions(createdBlockchain)

	createdBlockchainID = createdBlockchain.ID

	time.Sleep(1 * time.Second) // need time for cache refresh

	/* Get One Blockchain -> GET /blockchain/{id} */
	createdBlockchain, err = get[repository.Blockchain](fmt.Sprintf("blockchain/%s", createdBlockchainID))
	t.NoError(err)
	t.blockchainAssertions(createdBlockchain)

	/* Get All Blockchains -> GET /blockchain */
	createdBlockchains, err := get[[]repository.Blockchain]("blockchain")
	t.NoError(err)
	t.Len(createdBlockchains, 1)
	t.blockchainAssertions(createdBlockchains[0])

	/* Check Records Exist in Postgres DB as well as PHD Cache */
	pgBlockchains, err := t.PGDriver.ReadBlockchains()
	t.NoError(err)
	t.Len(pgBlockchains, 1)

	/* Activate Blockchain -> POST /blockchain/{id}/activate */
	blockchainActivated, err := post[bool](fmt.Sprintf("blockchain/%s/activate", createdBlockchainID), []byte("true"))
	t.NoError(err)
	t.True(blockchainActivated)

	time.Sleep(1 * time.Second) // need time for cache refresh

	activatedBlockchain, err := get[repository.Blockchain](fmt.Sprintf("blockchain/%s", createdBlockchainID))
	t.NoError(err)
	t.Equal(true, activatedBlockchain.Active)

	/* ERROR - Create Blockchain (duplicate record) -> POST /blockchain */
	_, err = post[repository.Blockchain]("blockchain", []byte(blockchainJSON))
	t.Equal("Response not OK. Internal Server Error", err.Error())

	/* ERROR - Create Blockchain (bad data) -> POST /blockchain */
	_, err = post[repository.Blockchain]("blockchain", []byte(`{"badJSON": "y tho",}`))
	t.Equal("Response not OK. Bad Request", err.Error())

	/* ERROR - Get One Blockchain (non-existent ID) -> GET /blockchain/{id} */
	_, err = get[repository.Blockchain](fmt.Sprintf("blockchain/%s", "NOT-REAL"))
	t.Equal("Response not OK. Not Found", err.Error())
}

func (t *PHDTestSuite) blockchainAssertions(blockchain repository.Blockchain) {
	t.NotEmpty(blockchain)
	t.Equal("TST01", blockchain.ID)
	t.Equal("https://test.external.com/rpc", blockchain.Altruist)
	t.Empty(blockchain.Redirects)
	t.Equal("TEST01", blockchain.Ticker)
	t.Equal("Test Mainnet", blockchain.Description)
	t.Equal(false, blockchain.Active)
	t.Equal("{\"method\":\"test_blockNumber\",\"id\":1,\"jsonrpc\":\"2.0\"}", blockchain.SyncCheckOptions.Body)
	t.Equal("test-mainnet", blockchain.BlockchainAliases[0])
}

func (t *PHDTestSuite) TestPHD_ApplicationEndpoints() {
	/* Create Application -> POST /application */
	createdApplication, err := post[repository.Application]("application", []byte(applicationJSON))
	t.NoError(err)
	t.applicationAssertions(createdApplication)

	createdApplicationID = createdApplication.ID

	time.Sleep(1 * time.Second) // need time for cache refresh

	/* Get One Application -> GET /application/{id} */
	createdApplication, err = get[repository.Application](fmt.Sprintf("application/%s", createdApplicationID))
	t.NoError(err)
	t.applicationAssertions(createdApplication)

	/* Get All Applications -> GET /application */
	createdApplications, err := get[[]repository.Application]("application")
	t.NoError(err)
	t.Len(createdApplications, 1)
	t.applicationAssertions(createdApplications[0])

	/* Get All of One User's Applications-> GET /user/{id}/application */
	userApplications, err := get[[]repository.Application](fmt.Sprintf("user/%s/application", testUserID))
	t.NoError(err)
	t.Len(userApplications, 1)
	t.applicationAssertions(userApplications[0])

	/* Check Records Exist in Postgres DB as well as PHD Cache */
	pgApplications, err := t.PGDriver.ReadApplications()
	t.NoError(err)
	t.Len(pgApplications, 1)

	/* Update One Application -> PUT /application/{id} */
	update := repository.UpdateApplication{
		Name:                 "update-application-1",
		NotificationSettings: &repository.NotificationSettings{Half: true, ThreeQuarters: false},
		GatewaySettings: &repository.GatewaySettings{
			SecretKeyRequired:    true,
			WhitelistOrigins:     []string{"test-origins-1", "test-origins-2"},
			WhitelistUserAgents:  []string{"test-agents-1", "test-agents-2", "test-agents-3"},
			WhitelistBlockchains: []string{"test-chains-1"},
			WhitelistContracts: []repository.WhitelistContract{
				{BlockchainID: "TST01", Contracts: []string{"test-contract-1"}},
			},
			WhitelistMethods: []repository.WhitelistMethod{
				{BlockchainID: "TST01", Methods: []string{"test-method-1"}},
				{BlockchainID: "TST01", Methods: []string{"test-method-2", "test-method-3"}},
			},
		},
	}
	updateJSON, err := json.Marshal(update)
	t.NoError(err)

	updatedApplication, err := put[repository.Application](fmt.Sprintf("application/%s", createdApplicationID), updateJSON)
	t.NoError(err)
	t.Equal("update-application-1", updatedApplication.Name)
	t.Equal(repository.PayPlanType("FREETIER_V0"), updatedApplication.Limit.PayPlan.Type)
	t.Equal(250000, updatedApplication.DailyLimit())
	t.Equal(true, updatedApplication.NotificationSettings.Half)
	t.Equal(false, updatedApplication.NotificationSettings.ThreeQuarters)
	t.Equal(true, updatedApplication.GatewaySettings.SecretKeyRequired)
	t.Len(updatedApplication.GatewaySettings.WhitelistOrigins, 2)
	t.Equal("test-origins-2", updatedApplication.GatewaySettings.WhitelistOrigins[1])
	t.Len(updatedApplication.GatewaySettings.WhitelistUserAgents, 3)
	t.Equal("test-agents-3", updatedApplication.GatewaySettings.WhitelistUserAgents[2])
	t.Len(updatedApplication.GatewaySettings.WhitelistBlockchains, 1)
	t.Equal("test-chains-1", updatedApplication.GatewaySettings.WhitelistBlockchains[0])
	t.Len(updatedApplication.GatewaySettings.WhitelistContracts, 1)
	t.Equal("TST01", updatedApplication.GatewaySettings.WhitelistContracts[0].BlockchainID)
	t.Equal("test-contract-1", updatedApplication.GatewaySettings.WhitelistContracts[0].Contracts[0])
	t.Len(updatedApplication.GatewaySettings.WhitelistMethods, 2)
	t.Equal("TST01", updatedApplication.GatewaySettings.WhitelistMethods[0].BlockchainID)
	t.Len(updatedApplication.GatewaySettings.WhitelistMethods[1].Methods, 2)
	t.Equal("test-method-3", updatedApplication.GatewaySettings.WhitelistMethods[1].Methods[1])
	t.NotEmpty(updatedApplication.UpdatedAt)

	/* Update One Application Pay Plan -> PUT /application/{id} */
	updatePayPlan := repository.UpdateApplication{
		Limit: &repository.AppLimit{
			PayPlan: repository.PayPlan{Type: repository.PayAsYouGoV0},
		},
	}
	updatePayPlanJSON, err := json.Marshal(updatePayPlan)
	t.NoError(err)

	updatedApplication, err = put[repository.Application](fmt.Sprintf("application/%s", createdApplicationID), updatePayPlanJSON)
	t.NoError(err)
	t.Equal("update-application-1", updatedApplication.Name)
	t.Equal(repository.PayPlanType("PAY_AS_YOU_GO_V0"), updatedApplication.Limit.PayPlan.Type)
	t.Equal(0, updatedApplication.DailyLimit())
	t.Equal(true, updatedApplication.NotificationSettings.Half)

	/* Update One Application Pay Plan to Enterprise (with custom limit) -> PUT /application/{id} */
	updateEnterprise := repository.UpdateApplication{
		Limit: &repository.AppLimit{
			PayPlan:     repository.PayPlan{Type: repository.Enterprise},
			CustomLimit: 4200000,
		},
	}
	updateEnterpriseJSON, err := json.Marshal(updateEnterprise)
	t.NoError(err)

	updatedApplication, err = put[repository.Application](fmt.Sprintf("application/%s", createdApplicationID), updateEnterpriseJSON)
	t.NoError(err)
	t.Equal("update-application-1", updatedApplication.Name)
	t.Equal(repository.PayPlanType("ENTERPRISE"), updatedApplication.Limit.PayPlan.Type)
	t.Equal(4200000, updatedApplication.DailyLimit())
	t.Equal("test-chains-1", updatedApplication.GatewaySettings.WhitelistBlockchains[0])

	/* Update First Date Surpassed -> POST /application/first_date_surpassed */
	updateDate := repository.UpdateFirstDateSurpassed{
		ApplicationIDs:     []string{createdApplication.ID},
		FirstDateSurpassed: time.Now(),
	}
	updateDateJSON, err := json.Marshal(updateDate)
	t.NoError(err)

	updatedDateApplication, err := post[[]repository.Application]("application/first_date_surpassed", updateDateJSON)
	t.NoError(err)
	t.NotEmpty(updatedDateApplication[0].FirstDateSurpassed)

	/* Get All Application Limits -> GET /application/limits */
	applicationLimits, err := get[[]repository.AppLimits]("application/limits")
	t.NoError(err)
	t.Len(applicationLimits, 1)
	t.Equal("update-application-1", applicationLimits[0].AppName)
	t.Equal(testUserID, applicationLimits[0].AppUserID)
	t.Equal("test_key_7a7d163434b10803eece4ddb2e0726e39ec6bb99b828aa309d05ffd", applicationLimits[0].PublicKey)
	t.Equal(repository.PayPlanType("ENTERPRISE"), applicationLimits[0].PlanType)
	t.Equal(4200000, applicationLimits[0].DailyLimit)
	t.Equal(true, applicationLimits[0].NotificationSettings.Half)
	t.Equal(false, applicationLimits[0].NotificationSettings.ThreeQuarters)
	t.NotEmpty(applicationLimits[0].FirstDateSurpassed)

	/* Remove One Application -> PUT /application/{id} (with Remove: true) */
	remove := repository.UpdateApplication{Remove: true}
	removeJSON, err := json.Marshal(remove)
	t.NoError(err)

	removedApplication, err := put[repository.Application](fmt.Sprintf("application/%s", createdApplicationID), removeJSON)
	t.NoError(err)
	t.Equal(repository.AppStatus("AWAITING_GRACE_PERIOD"), removedApplication.Status)

	/* ERROR - Create Application (bad data) -> POST /application */
	_, err = post[repository.Application]("application", []byte(`{"badJSON": "y tho",}`))
	t.Equal("Response not OK. Bad Request", err.Error())

	/* ERROR - Get One Application (non-existent ID) -> GET /application/{id} */
	_, err = get[repository.Application](fmt.Sprintf("application/%s", "not-a-real-id"))
	t.Equal("Response not OK. Not Found", err.Error())

	/* ERROR - Attempting to update non-Enterprise plan with custom limit -> PUT /application/{id} */
	updateEnterpriseErr := repository.UpdateApplication{
		Limit: &repository.AppLimit{
			PayPlan:     repository.PayPlan{Type: repository.FreetierV0},
			CustomLimit: 123456,
		},
	}
	updateEnterpriseErrJSON, err := json.Marshal(updateEnterpriseErr)
	t.NoError(err)

	updatedApplication, err = put[repository.Application](fmt.Sprintf("application/%s", createdApplicationID), updateEnterpriseErrJSON)
	t.Equal("Response not OK. Unprocessable Entity", err.Error())
}

func (t *PHDTestSuite) applicationAssertions(app repository.Application) {
	t.NotEmpty(app)
	t.NotEmpty(app.ID)
	t.Equal(testUserID, app.UserID)
	t.Equal("test-application-1", app.Name)
	t.Equal(repository.AppStatus("IN_SERVICE"), app.Status)
	t.Equal(true, app.Dummy)
	t.Equal(repository.PayPlanType("FREETIER_V0"), app.Limit.PayPlan.Type)
	t.Equal(250000, app.DailyLimit())
	t.Equal("test_address_8dbb89278918da056f589086fb4", app.GatewayAAT.Address)
	t.Equal("test_key_7a7d163434b10803eece4ddb2e0726e39ec6bb99b828aa309d05ffd", app.GatewayAAT.ApplicationPublicKey)
	t.Equal("test_key_f9c21a35787c53c8cdb532cad0dc01e099f34c28219528e3732c2da38037a84db4ce0282fa9aa404e56248155a1fbda789c8b5976711ada8588ead5", app.GatewayAAT.ApplicationSignature)
	t.Equal("test_key_2381d403a2e2edeb37c284957fb0ee5d66f1081acb87478a5817919", app.GatewayAAT.ClientPublicKey)
	t.Equal("test_key_0c0fbd26d98bcbdca4d4f14fdc45ffb6db0e6a23a56671fc4b444e1b8bbd4a934adde984d117f281867cb71d9537de45473b3028ead2326cd9e27ad", app.GatewayAAT.PrivateKey)
	t.Equal("0.0.1", app.GatewayAAT.Version)
	t.Equal("test_key_ba2724be652eca0a350bc07", app.GatewaySettings.SecretKey)
	t.Equal(false, app.GatewaySettings.SecretKeyRequired)
	t.Equal(false, app.NotificationSettings.Half)
	t.Equal(true, app.NotificationSettings.ThreeQuarters)
	t.Empty(app.FirstDateSurpassed)
	t.NotEmpty(app.CreatedAt)
	t.NotEmpty(app.UpdatedAt)
}

func (t *PHDTestSuite) TestPHD_LoadBalancerEndpoints() {
	/* Create Load Balancer -> POST /application */
	loadBalancerInput := []byte(fmt.Sprintf(loadBalancerJSON, createdApplicationID))

	createdLoadBalancer, err := post[repository.LoadBalancer]("load_balancer", loadBalancerInput)
	t.NoError(err)
	t.loadBalancerAssertions(createdLoadBalancer)

	time.Sleep(1 * time.Second) // need time for cache refresh

	/* Get One Load Balancer -> GET /load_balancer/{id} */
	createdLoadBalancer, err = get[repository.LoadBalancer](fmt.Sprintf("load_balancer/%s", createdLoadBalancer.ID))
	t.NoError(err)
	t.loadBalancerAssertions(createdLoadBalancer)

	/* Get All Load Balancers -> GET /load_balancer */
	createdLoadBalancers, err := get[[]repository.LoadBalancer]("load_balancer")
	t.NoError(err)
	t.Len(createdLoadBalancers, 1)
	t.loadBalancerAssertions(createdLoadBalancers[0])

	/* Get All of One User's Load Balancers -> GET /user/{id}/load_balancer */
	userLoadBalancers, err := get[[]repository.LoadBalancer](fmt.Sprintf("user/%s/load_balancer", testUserID))
	t.NoError(err)
	t.Len(userLoadBalancers, 1)
	t.loadBalancerAssertions(userLoadBalancers[0])

	/* Check Records Exist in Postgres DB as well as PHD Cache */
	pgLoadBalancers, err := t.PGDriver.ReadLoadBalancers()
	t.NoError(err)
	t.Len(pgLoadBalancers, 1)

	/* Update One Load Balancer -> PUT /load_balancer/{id} */
	update := repository.UpdateLoadBalancer{
		Name: "update-load-balancer-1",
		StickyOptions: &repository.StickyOptions{
			Duration:      "test-duration",
			StickyOrigins: []string{"test-origins-1", "test-origins-2"},
			StickyMax:     200,
			Stickiness:    true,
		},
	}
	updateJSON, err := json.Marshal(update)
	t.NoError(err)

	updatedLoadBalancer, err := put[repository.LoadBalancer](fmt.Sprintf("load_balancer/%s", createdLoadBalancer.ID), updateJSON)
	t.NoError(err)
	t.Equal("update-load-balancer-1", updatedLoadBalancer.Name)
	t.Equal("test-duration", updatedLoadBalancer.StickyOptions.Duration)
	t.Len(updatedLoadBalancer.StickyOptions.StickyOrigins, 2)
	t.Equal("test-origins-2", updatedLoadBalancer.StickyOptions.StickyOrigins[1])
	t.Equal(200, updatedLoadBalancer.StickyOptions.StickyMax)
	t.Equal(true, updatedLoadBalancer.StickyOptions.Stickiness)

	/* Remove One Load Balancer -> PUT /load_balancer/{id} (with Remove: true) */
	remove := repository.UpdateLoadBalancer{Remove: true}
	removeJSON, err := json.Marshal(remove)
	t.NoError(err)

	removedLoadBalancer, err := put[repository.LoadBalancer](fmt.Sprintf("load_balancer/%s", createdLoadBalancer.ID), removeJSON)
	t.NoError(err)
	t.Equal("", removedLoadBalancer.UserID)

	/* ERROR - Create Load Balancer (bad data) -> POST /load_balancer */
	_, err = post[repository.LoadBalancer]("load_balancer", []byte(`{"badJSON": "y tho",}`))
	t.Equal("Response not OK. Bad Request", err.Error())

	/* ERROR - Get One Load Balancer (non-existent ID) -> GET /load_balancer/{id} */
	_, err = get[repository.LoadBalancer](fmt.Sprintf("load_balancer/%s", "not-a-real-id"))
	t.Equal("Response not OK. Not Found", err.Error())
}

func (t *PHDTestSuite) loadBalancerAssertions(lb repository.LoadBalancer) {
	t.NotEmpty(lb)
	t.NotEmpty(lb.ID)
	t.Equal("test-load-balancer-1", lb.Name)
	t.Equal(testUserID, lb.UserID)
	t.Equal([]string(nil), lb.ApplicationIDs)
	t.Equal(2000, lb.RequestTimeout)
	t.Equal(false, lb.Gigastake)
	t.Equal(true, lb.GigastakeRedirect)
	t.Equal("", lb.StickyOptions.Duration)
	t.Equal([]string(nil), lb.StickyOptions.StickyOrigins)
	t.Equal(0, lb.StickyOptions.StickyMax)
	t.Equal(false, lb.StickyOptions.Stickiness)
	t.Len(lb.Applications, 1)
	t.Equal(createdApplicationID, lb.Applications[0].ID)
	t.Equal("update-application-1", lb.Applications[0].Name)
	t.NotEmpty(lb.CreatedAt)
	t.NotEmpty(lb.UpdatedAt)
}

func (t *PHDTestSuite) TestPHD_PayPlanEndpoints() {
	/* Get All Pay Plans -> GET /pay_plan */
	payPlans, err := get[[]repository.PayPlan]("pay_plan")
	t.NoError(err)
	t.Len(payPlans, 6)

	/* Get One Pay Plan -> GET /pay_plan/{type} */
	payPlan, err := get[repository.PayPlan](fmt.Sprintf("pay_plan/%s", "FREETIER_V0"))
	t.NoError(err)
	t.Equal(repository.PayPlanType("FREETIER_V0"), payPlan.Type)
	t.Equal(250000, payPlan.Limit)

	/* Check Records Exist in Postgres DB as well as PHD Cache */
	pgPayPlans, err := t.PGDriver.ReadPayPlans()
	t.NoError(err)
	t.Len(pgPayPlans, 6)

	/* ERROR - Get One Pay Plan (non-existent ID) -> GET /pay_plan/{type} */
	_, err = get[repository.PayPlan](fmt.Sprintf("pay_plan/%s", "not-a-real-pay-plan"))
	t.Equal("Response not OK. Not Found", err.Error())
}

func (t *PHDTestSuite) TestPHD_RedirectEndpoints() {
	redirectInput := []byte(fmt.Sprintf(redirectJSON, createdBlockchainID))

	/* Create Redirect -> POST /redirect */
	createdRedirect, err := post[repository.Redirect]("redirect", redirectInput)
	t.NoError(err)
	t.Equal(createdBlockchainID, createdRedirect.BlockchainID)
	t.Equal("test-mainnet", createdRedirect.Alias)
	t.Equal("test-rpc.gateway.pokt.network", createdRedirect.Domain)
	t.Equal("12345", createdRedirect.LoadBalancerID)

	/* Check Records Exist in Postgres DB as well as PHD Cache */
	pgRedirects, err := t.PGDriver.ReadRedirects()
	t.NoError(err)
	t.Len(pgRedirects, 1)

	/* ERROR - Create Redirect (duplicate record) -> POST /redirect */
	_, err = post[repository.Redirect]("redirect", []byte(redirectJSON))
	t.Equal("Response not OK. Internal Server Error", err.Error())

	/* ERROR - Create Redirect (bad data) -> POST /redirect */
	_, err = post[repository.Redirect]("redirect", []byte(`{"badJSON": "y tho",}`))
	t.Equal("Response not OK. Bad Request", err.Error())
}

/* Test Client HTTP Funcs */
func get[T any](path string) (T, error) {
	rawURL := fmt.Sprintf("%s/%s", baseURL, path)

	headers := http.Header{"Authorization": {apiKey}}

	var data T

	response, err := testClient.Get(rawURL, headers)
	if err != nil {
		return data, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return data, fmt.Errorf("%w. %s", ErrResponseNotOK, http.StatusText(response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func post[T any](path string, postData []byte) (T, error) {
	var data T

	rawURL := fmt.Sprintf("%s/%s", baseURL, path)

	headers := http.Header{
		"Authorization": {apiKey},
		"Content-Type":  {"application/json"},
		"Connection":    {"Close"},
	}

	postBody := bytes.NewBufferString(string(postData))

	response, err := testClient.Post(rawURL, postBody, headers)
	if err != nil {
		return data, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return data, fmt.Errorf("%w. %s", ErrResponseNotOK, http.StatusText(response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func put[T any](path string, postData []byte) (T, error) {
	var data T

	rawURL := fmt.Sprintf("%s/%s", baseURL, path)

	headers := http.Header{
		"Authorization": {apiKey},
		"Content-Type":  {"application/json"},
		"Connection":    {"Close"},
	}

	postBody := bytes.NewBufferString(string(postData))

	response, err := testClient.Put(rawURL, postBody, headers)
	if err != nil {
		return data, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return data, fmt.Errorf("%w. %s", ErrResponseNotOK, http.StatusText(response.StatusCode))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}
