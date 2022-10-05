package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	apiKey  = "bwLzAhU2GBnHKe59Lqax"
)

var (
	// ErrNoUserID         error = errors.New("No User ID")
	// ErrNoApplicationID  error = errors.New("No Application ID")
	// ErrNoLoadBalancerID error = errors.New("No Load Balancer ID")
	ErrResponseNotOK error = errors.New("Response not OK")

	testClient = httpclient.NewClient(httpclient.WithHTTPTimeout(5*time.Second), httpclient.WithRetryCount(0))
)

type (
	PHDTestSuite struct{ suite.Suite }
)

func (t *PHDTestSuite) SetupSuite() {
	_, err := exec.Command("docker", "compose", "-f", "./docker-compose.yml", "up", "-d").Output()
	t.NoError(err)
}

func (t *PHDTestSuite) TearDownSuite() {
	_, err := exec.Command("docker-compose", "-f", "./docker-compose.yml", "down", "--remove-orphans", "--rmi", "all", "-v").Output()
	t.NoError(err)
}

func TestE2E_RunSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end to end test")
	}

	suite.Run(t, new(PHDTestSuite))
}

/*
The End-to-End test suite builds a container from the Pocket HTTP DB Docker image that we deploy,
and connects to a standard Postgres container initialized with the database tables used in production.

The test then performs every operation that PHD can perform for each set of endpoints and checks results.
*/

func (t *PHDTestSuite) TestPHD_BlockchainEndpoints() {
	/* Create Blockchain -> POST /blockchain */
	blockchain, err := ioutil.ReadFile("./blockchain.json")
	t.NoError(err)

	createdBlockchain, err := post[repository.Blockchain]("blockchain", blockchain)
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
	t.Equal("test-mainnet", blockchain.Redirects[0].Alias)
	t.Equal("test-rpc.gateway.pokt.network", blockchain.Redirects[0].Domain)
	t.Equal("TEST01", blockchain.Ticker)
	t.Equal("Test Mainnet", blockchain.Description)
	t.Equal(false, blockchain.Active)
	t.Equal("{\"method\":\"test_blockNumber\",\"id\":1,\"jsonrpc\":\"2.0\"}", blockchain.SyncCheckOptions.Body)
	t.Equal("test-mainnet", blockchain.BlockchainAliases[0])
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
