package timeseries

import (
	"errors"
	"net/http"

	"timeengine/namespace"
	"timeengine/users"

	"appengine"
	"appengine/user"

	"third_party/go-endpoints/endpoints"
)

const clientIdUser = "731315961832-2s0uos83jp6mtqmsgnh21d198871vjnl.apps.googleusercontent.com"
const clientIdRobot = "731315961832-2dtj4qa9bma3nfkr70nv64thpa1ndub4.apps.googleusercontent.com"

var (
	scopes = []string{endpoints.EmailScope}
	clientIds = []string{endpoints.ApiExplorerClientId, clientIdUser, clientIdRobot}
	audiences = []string{clientIdUser, clientIdRobot}
)

type PutResp struct{}
type DataPointService struct{}

func (dps *DataPointService) Put(r *http.Request, req *PutReq, resp *PutResp) error {
	epCtx := endpoints.NewContext(r)
	user, err := getCurrentUser(epCtx)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("Not authentified")
	}

	c := appengine.NewContext(r)
	// Force the user to be a known user. i.e. someone who already have access to the
	// app.
	if ok, _, _ := users.IsUserAuthorized(r, user); !ok {
		return errors.New("Unknown user " + user.Email)
	}

	if !namespace.VerifyNamespace(c, req.Ns, req.NsSecret) {
		return errors.New("Missing or unknown namespace/secret")
	}
	delayInputProcess.Call(c, req)
	return nil
}

func getCurrentUser(c endpoints.Context) (*user.User, error) {
	u, err := endpoints.CurrentUser(c, scopes, audiences, clientIds)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, errors.New("Unauthorized: Please, sign in.")
	}
	c.Debugf("Current user: %#v", u)
	return u, nil
}

func RegisterService() (*endpoints.RpcService, error) {
	dataPointService := &DataPointService{}
	api, err := endpoints.RegisterService(dataPointService, "timeengine", "v1", "DataPoints API", true)
	if err != nil {
		panic(err.Error())
	}

	info := api.MethodByName("Put").Info()
	//info.Name = "datapoints.put"
	//info.HttpMethod = "GET"
	//info.Path = "putdatapoints"
	info.Desc = "Push data points."
	info.Scopes = scopes
	info.ClientIds = clientIds
	return api, nil
}
