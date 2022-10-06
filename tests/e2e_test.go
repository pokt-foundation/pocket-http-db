package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/gojektech/heimdall/httpclient"
	"github.com/pokt-foundation/portal-api-go/repository"
	"github.com/stretchr/testify/suite"
)

const (
	baseURL = "http://localhost:8080"
	apiKey  = "test_api_key_6789"

	testUserID = "test_id_de26a0db3b6c631c4"
)

var (
	// ErrNoUserID         error = errors.New("No User ID")
	// ErrNoApplicationID  error = errors.New("No Application ID")
	// ErrNoLoadBalancerID error = errors.New("No Load Balancer ID")
	ErrResponseNotOK error = errors.New("Response not OK")

	testClient = httpclient.NewClient(httpclient.WithHTTPTimeout(5*time.Second), httpclient.WithRetryCount(0))

	createdApplicationID string = "" // Used to create a LoadBalancer.ApplicationIDs slice
)

type (
	PHDTestSuite struct{ suite.Suite }
)

func (t *PHDTestSuite) SetupSuite() {
	output, err := exec.Command("docker", "compose", "up", "-d").Output()
	fmt.Println("TESTING STARTUP", string(output), err)
	t.NoError(err)
}

func (t *PHDTestSuite) TearDownSuite() {
	output, err := exec.Command("docker-compose", "down", "--remove-orphans", "-v").Output()
	fmt.Println("TESTING TEARDOWN", string(output), err)
	t.NoError(err)
}

func TestE2E_RunSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end to end test")
	}

	output, err := exec.Command("docker", "ps", "-a").Output()
	fmt.Println("TESTING DEBUG", string(output), err)

	suite.Run(t, new(PHDTestSuite))
}

/*
The End-to-End test suite builds a container from the Pocket HTTP DB Docker image that we deploy,
and connects to a standard Postgres container initialized with the database tables used in production.

The test then performs every operation that PHD can perform for each set of endpoints and checks results.

TODO:
1. Add error testing (send bad data, try and create duplicate records, ...)
2. Finish Load Balancer Endpoints test
3. Test the effect of setting the cache
*/

func (t *PHDTestSuite) TestPHD_BlockchainEndpoints() {
	/* Create Blockchain -> POST /blockchain */
	createdBlockchain, err := post[repository.Blockchain]("blockchain", []byte(blockchainJSON))
	t.NoError(err)
	t.blockchainAssertions(createdBlockchain)

	/* Get One Blockchain -> GET /blockchain/{id} */
	createdBlockchain, err = get[repository.Blockchain](fmt.Sprintf("blockchain/%s", createdBlockchain.ID))
	t.NoError(err)
	t.blockchainAssertions(createdBlockchain)

	/* Get All Blockchains -> GET /blockchain */
	createdBlockchains, err := get[[]repository.Blockchain]("blockchain")
	t.NoError(err)
	t.Len(createdBlockchains, 1)
	t.blockchainAssertions(createdBlockchains[0])

	/* Activate Blockchain -> POST /blockchain/{id}/activate */
	blockchainActivated, err := post[bool](fmt.Sprintf("blockchain/%s/activate", createdBlockchain.ID), []byte("true"))
	t.NoError(err)
	t.True(blockchainActivated)

	activatedBlockchain, err := get[repository.Blockchain](fmt.Sprintf("blockchain/%s", createdBlockchain.ID))
	t.NoError(err)
	t.Equal(true, activatedBlockchain.Active)
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
	t.NotEmpty(blockchain.CreatedAt)
	t.NotEmpty(blockchain.UpdatedAt)
}

func (t *PHDTestSuite) TestPHD_ApplicationEndpoints() {
	/* Create Application -> POST /application */
	createdApplication, err := post[repository.Application]("application", []byte(applicationJSON))
	t.NoError(err)
	t.applicationAssertions(createdApplication)

	createdApplicationID = createdApplication.ID

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

	/* Update One Application -> PUT /application/{id} */
	update := repository.UpdateApplication{
		Name:                 "update-application-1",
		PayPlanType:          "PAY_AS_YOU_GO_V0",
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
	t.Equal(repository.PayPlanType(""), updatedApplication.PayPlanType)
	t.Equal(repository.PayPlanType("PAY_AS_YOU_GO_V0"), updatedApplication.Limits.PlanType)
	t.Equal(0, updatedApplication.Limits.DailyLimit)
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
	t.Equal(repository.PayPlanType("PAY_AS_YOU_GO_V0"), applicationLimits[0].PlanType)
	t.Equal(0, applicationLimits[0].DailyLimit)
	t.Equal(true, applicationLimits[0].NotificationSettings.Half)
	t.Equal(false, applicationLimits[0].NotificationSettings.ThreeQuarters)
	t.NotEmpty(applicationLimits[0].FirstDateSurpassed)
}

func (t *PHDTestSuite) applicationAssertions(app repository.Application) {
	t.NotEmpty(app)
	t.NotEmpty(app.ID)
	t.Equal(testUserID, app.UserID)
	t.Equal("test-application-1", app.Name)
	t.Equal(repository.AppStatus("IN_SERVICE"), app.Status)
	t.Equal(true, app.Dummy)
	t.Equal(repository.PayPlanType(""), app.PayPlanType)
	t.Equal(repository.PayPlanType("FREETIER_V0"), app.Limits.PlanType)
	t.Equal(250000, app.Limits.DailyLimit)
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

// TODO - Finish Load Balancer Endpoint Tests
// func (t *PHDTestSuite) TestPHD_LoadBalancerEndpoints() {
// 	/* Create Load Balancer -> POST /application */
// 	loadBalancerInput := []byte(fmt.Sprintf(loadBalancerJSON, createdApplicationID))

// 	createdLoadBalancer, err := post[repository.LoadBalancer]("load_balancer", loadBalancerInput)
// 	t.NoError(err)
// 	t.loadBalancerAssertions(createdLoadBalancer)

// 	/* Get One Load Balancer -> GET /load_balancer/{id} */
// 	createdLoadBalancer, err = get[repository.LoadBalancer](fmt.Sprintf("load_balancer/%s", createdLoadBalancer.ID))
// 	t.NoError(err)
// 	t.loadBalancerAssertions(createdLoadBalancer)

// 	/* Get All Load Balancers -> GET /load_balancer */
// 	createdLoadBalancers, err := get[[]repository.LoadBalancer]("load_balancer")
// 	t.NoError(err)
// 	t.Len(createdLoadBalancers, 1)
// 	t.loadBalancerAssertions(createdLoadBalancers[0])

// 	/* Get All of One User's Load Balancers -> GET /user/{id}/load_balancer */
// 	userLoadBalancers, err := get[[]repository.LoadBalancer](fmt.Sprintf("user/%s/load_balancer", testUserID))
// 	t.NoError(err)
// 	t.Len(userLoadBalancers, 1)
// 	t.loadBalancerAssertions(userLoadBalancers[0])

// 	/* Update One Load Balancer -> PUT /load_balancer/{id} */
// }

// func (t *PHDTestSuite) loadBalancerAssertions(lb repository.LoadBalancer) {
// 	t.NotEmpty(lb)
// 	t.NotEmpty(lb.ID)
// 	t.Equal("test-load-balancer-1", lb.Name)
// 	t.Equal(testUserID, lb.UserID)
// 	t.Equal([]string{"test_app_id_47fht6s5fd62"}, lb.ApplicationIDs)
// 	t.Equal(2000, lb.RequestTimeout)
// 	t.Equal(false, lb.Gigastake)
// 	t.Equal(true, lb.GigastakeRedirect)
// 	t.Equal("", lb.StickyOptions.Duration)
// 	t.Equal([]string{}, lb.StickyOptions.StickyOrigins)
// 	t.Equal(0, lb.StickyOptions.StickyMax)
// 	t.Equal(false, lb.StickyOptions.Stickiness)
// 	t.Len(lb.Applications, 1)
// 	t.Len(lb.Applications[0].Name, 1)
// 	t.NotEmpty(lb.CreatedAt)
// 	t.NotEmpty(lb.UpdatedAt)
// }

func (t *PHDTestSuite) TestPHD_PayPlanEndpoints() {
	/* Get All Pay Plans -> GET /pay_plan */
	payPlans, err := get[[]repository.PayPlan]("pay_plan")
	t.NoError(err)
	t.Len(payPlans, 5)

	/* Get One PayPlan -> GET /pay_plan/{type} */
	payPlan, err := get[repository.PayPlan](fmt.Sprintf("pay_plan/%s", "FREETIER_V0"))
	t.NoError(err)
	t.Equal(repository.PayPlanType("FREETIER_V0"), payPlan.PlanType)
	t.Equal(250000, payPlan.DailyLimit)
}

func (t *PHDTestSuite) TestPHD_RedirectEndpoints() {
	/* Create Redirect -> POST /redirect */
	createdRedirect, err := post[repository.Redirect]("redirect", []byte(redirectJSON))

	t.NoError(err)
	t.Equal("TST01", createdRedirect.BlockchainID)
	t.Equal("test-mainnet", createdRedirect.Alias)
	t.Equal("test-rpc.gateway.pokt.network", createdRedirect.Domain)
	t.Equal("12345", createdRedirect.LoadBalancerID)
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

const (
	blockchainJSON = `{
		"id": "TST01",
		"altruist": "https://test.external.com/rpc",
		"redirects": [],
		"ticker": "TEST01",
		"chainID": "10",
		"network": "TST-01",
		"description": "Test Mainnet",
		"blockchain": "test-mainnet",
		"active": false,
		"enforceResult": "JSON",
		"logLimitBlocks": 100000,
		"appCount": 0,
		"syncCheckOptions": {
			"blockchainID": "TST01",
			"body": "{\"method\":\"test_blockNumber\",\"id\":1,\"jsonrpc\":\"2.0\"}",
			"resultKey": "result",
			"allowance": 3,
			"path": null
		},
		"blockchainAliases": ["test-mainnet"]
	}`

	applicationJSON = `{
		"id": "",
		"userID": "test_id_de26a0db3b6c631c4",
		"name": "test-application-1",
		"contactEmail": "",
		"description": "",
		"owner": "",
		"url": "",
		"status": "IN_SERVICE",
		"dummy": true,
		"payPlanType": "FREETIER_V0",
		"firstDateSurpassed": null,
		"gatewayAAT": {
			"address": "test_address_8dbb89278918da056f589086fb4",
			"applicationPublicKey": "test_key_7a7d163434b10803eece4ddb2e0726e39ec6bb99b828aa309d05ffd",
			"applicationSignature": "test_key_f9c21a35787c53c8cdb532cad0dc01e099f34c28219528e3732c2da38037a84db4ce0282fa9aa404e56248155a1fbda789c8b5976711ada8588ead5",
			"clientPublicKey": "test_key_2381d403a2e2edeb37c284957fb0ee5d66f1081acb87478a5817919",
			"privateKey": "test_key_0c0fbd26d98bcbdca4d4f14fdc45ffb6db0e6a23a56671fc4b444e1b8bbd4a934adde984d117f281867cb71d9537de45473b3028ead2326cd9e27ad",
			"version": "0.0.1"
		},
		"gatewaySettings": {
			"secretKey": "test_key_ba2724be652eca0a350bc07",
			"secretKeyRequired": false
		},
			"notificationSettings": {
			"signedUp": true,
			"quarter": false,
			"half": false,
			"threeQuarters": true,
			"full": true
		},
		"limits": {
			"planType": "",
			"dailyLimit": 0
		}
	}`

	// loadBalancerJSON = `{
	// 	"id": "",
	// 	"name": "test-load-balancer-1",
	// 	"userID": "test_id_de26a0db3b6c631c4",
	// 	"applicationIDs": ["%s"],
	// 	"requestTimeout": 2000,
	// 	"gigastake": false,
	// 	"gigastakeRedirect": true,
	// 	"stickinessOptions": {
	// 		"duration": "",
	// 		"stickyOrigins": null,
	// 		"stickyMax": 0,
	// 		"stickiness": false
	// 	},
	// 	"Applications": null,
	// }`

	redirectJSON = `{
		"id": "",
		"blockchainID": "TST01",
		"alias": "test-mainnet",
		"domain": "test-rpc.gateway.pokt.network",
		"loadBalancerID": "12345"
	}`
)
