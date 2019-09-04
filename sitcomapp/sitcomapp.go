package sitcomapp

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/saguywalker/sitcomchain/code"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	"github.com/tendermint/tendermint/version"
	dbm "github.com/tendermint/tm-db"
)

var (
	_               types.Application = (*SITComApplication)(nil)
	ProtocolVersion version.Protocol  = 0x1
	stateKey        []byte            = []byte("stateKey:")
)

type SITComApplication struct {
	types.BaseApplication

	state              State
	ValUpdates         []types.ValidatorUpdate
	valAddrToPubKeyMap map[string]types.PubKey
	logger             log.Logger
}

type State struct {
	db      dbm.DB
	Size    uint64 `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

func loadState(db dbm.DB) State {
	stateBytes := db.Get(stateKey)
	var state State
	if len(stateBytes) != 0 {
		err := json.Unmarshal(stateBytes, &state)
		if err != nil {
			panic(err)
		}
	}
	state.db = db
	return state
}

func saveState(state State) {
	stateBytes, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	state.db.Set(stateKey, stateBytes)
}

func NewSITComApplication(dbDir string) *SITComApplication {
	name := "sitcomchain"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &SITComApplication{
		state:              state,
		valAddrToPubKeyMap: make(map[string]types.PubKey),
		logger:             log.NewNopLogger()}
}

func (app *SITComApplication) SetLogger(l log.Logger) {
	app.logger = l
}

func (app *SITComApplication) Info(req types.RequestInfo) types.ResponseInfo {
	res := types.ResponseInfo{
		Data:       fmt.Sprintf("{\"size\":%v}", app.state.Size),
		Version:    version.ABCIVersion,
		AppVersion: ProtocolVersion.Uint64(),
	}
	res.LastBlockHeight = app.state.Height
	res.LastBlockAppHash = app.state.AppHash
	fmt.Printf("%+v\n", res)
	return res
}

// val:${pubkey}!{0 or 1} => update validator
// "add_competence=\{\"student_id\":\"59130500211\",\"competence_id\":\"30001\",\"by\":\"$publickey\",\"semester\":"12019"\}"'
// "approve_activity=\{\"student_id\":\"59130500211\",\"activity_id\":\"4999999999\",\"approver\":\"$publickey\",\"semester\":\"12019\"\}"'
func (app *SITComApplication) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	fmt.Println("In DeliverTx")
	
	if isValidatorTx(req.Tx) {
		return app.execValidatorTx(req.Tx)
	}

	parts := bytes.Split(req.Tx, []byte("="))

	if len(parts) != 2 {
		return types.ResponseDeliverTx{Code: code.CodeTypeBadData}
	}

	returnLog := ""
	var returnEvents []types.Event
	returnCode := code.CodeTypeBadData

	switch string(parts[0]) {
	case "add_competence":
		events, err := app.StaffAddCompetence(parts[1])
		if err != nil {
			returnLog = fmt.Sprint("Error with adding competence:", err)
		} else {
			returnCode = code.CodeTypeOK
			returnEvents = events
		}
		break
	case "approve_activity":
		events, err := app.AttendedActivity(parts[1])
		if err != nil {
			returnLog = fmt.Sprint("Error with updating attended activity:", err)
		} else {
			returnCode = code.CodeTypeOK
			returnEvents = events
		}
		break
	default:
		returnLog = fmt.Sprintf("Unknown function '%s'", parts[1])
	}

	return types.ResponseDeliverTx{Code: returnCode, Log: returnLog, Events: returnEvents}
}

func (app *SITComApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	fmt.Printf("CheckTx: %+v\n", string(req.Tx))
	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (app *SITComApplication) Commit() types.ResponseCommit {
	appHash := make([]byte, 8)
	binary.LittleEndian.PutUint64(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}
}

func (app *SITComApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	if len(req.Data) == 0 {
		app.state.db.Print()
	} else {
		res.Key = req.Data
		value := app.state.db.Get(req.Data)
		res.Value = value
		if value != nil {
			res.Log = "exists"
		} else {
			res.Log = "does not exist"
		}
	}
	return
}

func (app *SITComApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

func (app *SITComApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]types.ValidatorUpdate, 0)

	for _, ev := range req.ByzantineValidators {
		if ev.Type == tmtypes.ABCIEvidenceTypeDuplicateVote {
			// decrease voting power by 1
			if ev.TotalVotingPower == 0 {
				continue
			}
			app.updateValidator(types.ValidatorUpdate{
				PubKey: app.valAddrToPubKeyMap[string(ev.Validator.Address)],
				Power:  ev.TotalVotingPower - 1,
			})
		}
	}
	return types.ResponseBeginBlock{}
}

func (app *SITComApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

// ---------------------------------------------------------------------------

func (app *SITComApplication) Validators() (validators []types.ValidatorUpdate) {
	itr := app.state.db.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(types.ValidatorUpdate)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	return
}

func MakeValSetChangeTx(pubkey types.PubKey, power int64) []byte {
	pubStr := base64.StdEncoding.EncodeToString(pubkey.Data)
	return []byte(fmt.Sprintf("val:%s!%d", pubStr, power))
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), "val:")
}

func (app *SITComApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len("val:"):]

	//get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), "!")
	if len(pubKeyAndPower) != 2 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey!power'. Got %v", pubKeyAndPower)}
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	// decode the pubkey
	pubkey, err := base64.StdEncoding.DecodeString(pubkeyS)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid base64", pubkeyS)}
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(types.Ed25519ValidatorUpdate(pubkey, int64(power)))
}

/*func byteToHex(input []byte) string {
	var hexValue string
	for _, v := range input {
		hexValue += fmt.Sprintf("%02x", v)
	}
	return hexValue
}*/

func (app *SITComApplication) updateValidator(v types.ValidatorUpdate) types.ResponseDeliverTx {
	key := []byte("val:" + string(v.PubKey.Data))

	pubkey := ed25519.PubKeyEd25519{}
	copy(pubkey[:], v.PubKey.Data)

	if v.Power == 0 {
		// remove validator
		if !app.state.db.Has(key) {
			pubStr := base64.StdEncoding.EncodeToString(v.PubKey.Data)
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %s", pubStr)}
		}
		app.state.db.Delete(key)
		delete(app.valAddrToPubKeyMap, string(pubkey.Address()))
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.state.db.Set(key, value.Bytes())
		app.valAddrToPubKeyMap[string(pubkey.Address())] = v.PubKey
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return types.ResponseDeliverTx{Code: code.CodeTypeOK}
}

//---------------- Tendermint ABCI ---------------------

type StaffAddCompetence struct {
	StudentID    string `json:"student_id"`
	CompetenceID string `json:"competence_id"`
	By           string `json:"by"`
	Semester     uint16 `json:"semester"`
	Nonce        uint64 `json:"nonce"`
}

type AttendedActivity struct {
	StudentID  string `json:"student_id"`
	ActivityID string `json:"activity_id"`
	Approver   []byte `json:"approver"`
	Semester   uint16 `json:"semester"`
	Nonce      uint64 `json:"nonce"`
}

func (app *SITComApplication) StaffAddCompetence(body []byte) ([]types.Event, error) {
	var update StaffAddCompetence
	err := json.Unmarshal(body, &update)
	if err != nil {
		return nil, err
	}

	update.Nonce = app.state.Size + 1

	value, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	semester := make([]byte, 2)
	binary.LittleEndian.PutUint16(semester, update.Semester)

	key := crypto.Sha256(value)

	//Set struct_id to value
	app.state.db.Set(key, value)
	app.state.Size++

	events := []types.Event{
		{
			Type: "competence.add",
			Attributes: []cmn.KVPair{
				{Key: []byte("txid"), Value: key},
				{Key: []byte("studentid"), Value: []byte(update.StudentID)},
				{Key: []byte("competenceid"), Value: []byte(update.CompetenceID)},
				{Key: []byte("by"), Value: []byte(update.By)},
				{Key: []byte("semester"), Value: semester},
			},
		},
	}

	return events, nil
}

func (app *SITComApplication) AttendedActivity(body []byte) ([]types.Event, error) {
	var update AttendedActivity
	err := json.Unmarshal(body, &update)
	if err != nil {
		return nil, err
	}

	update.Nonce = app.state.Size + 1

	value, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	semester := make([]byte, 2)
	binary.LittleEndian.PutUint16(semester, update.Semester)

	key := crypto.Sha256(value)

	app.state.db.Set(key, value)
	app.state.Size++

	events := []types.Event{
		{
			Type: "activity.approve",
			Attributes: []cmn.KVPair{
				{Key: []byte("txid"), Value: key},
				{Key: []byte("studentid"), Value: []byte(update.StudentID)},
				{Key: []byte("activityid"), Value: []byte(update.ActivityID)},
				{Key: []byte("by"), Value: update.Approver},
				{Key: []byte("semester"), Value: semester},
			},
		},
	}

	return events, nil
}
