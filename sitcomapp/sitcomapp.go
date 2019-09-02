package sitcomapp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/saguywalker/sitcomchain/code"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tm-db"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

var (
	_               types.Application = (*SITComApplication)(nil)
	stateKey                          = []byte("stateKey")
	kvPairPrefixKey                   = []byte("kvPairKey:")
	ProtocolVersion version.Protocol  = 0x1
)

type SITComApplication struct {
	types.BaseApplication

	state      State
	ValUpdates []types.ValidatorUpdate
	valAddrToPubKeyMap map[string]types.PubKey
	logger     log.Logger
}

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
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

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

func 

func NewSITComApplication(dbDir string) *SITComApplication {
	name := "sitcomchain"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &SITComApplication{
		state:  state,
		valAddrToPubKeyMap: make(map[string]types.PubKey),
		logger: log.NewNopLogger()}
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
	return res
}

func (app *SITComApplication) DeliverTx(req types.RequestDeliverTx) types.ResponseDeliverTx {
	if isValidatorTx(req.Tx) {										// val:${pubkey}!{0 or 1}
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
	case "auto_add_competence":
		events, err := app.CreateCompetence(parts[1])
		if err != nil {
			returnLog = fmt.Sprint("Error with creating competence:", err)
		} else {
			returnCode = code.CodeTypeOK
			returnEvents = events
		}
		break
	case "staff_add_competence":
		events, err := app.StaffAddCompetence(parts[1])
		if err != nil {
			returnLog = fmt.Sprint("Error with adding competence:", err)
		} else {
			returnCode = code.CodeTypeOK
			returnEvents = events
		}
		break
	case "attended_activity":
		events, err := app.AttendedActivity(parts[1])
		if err != nil {
			returnLog = fmt.Sprint("Error with updating attended activity:", err)
		} else {
			returnCode = code.CodeTypeOK
			returnEvents = events
		}
		break
	}

	return types.ResponseDeliverTx{Code: returnCode, Log: returnLog, Events: returnEvents}
}


func (app *SITComApplication) CheckTx(req types.RequestCheckTx) types.ResponseCheckTx {
	return types.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (app *SITComApplication) Commit() types.ResponseCommit {
	appHash := make([]byte, 8)
	binary.PutVarint(appHash, app.state.Size)
	app.state.AppHash = appHash
	app.state.Height++
	saveState(app.state)
	return types.ResponseCommit{Data: appHash}
}

/*
func (app *SITComApplication) Query(req types.RequestQuery) (res types.ResponseQuery) {
	if len(req.Data) == 0{
    //str, _ := app.state.db.db.GetProperty("leveldb.stats")
    //fmt.Printf("%v\n", str)
    iter := app.state.db.Iterator(nil, nil)
	  for ; iter.Valid(); iter.Next() {
		  key := iter.Key()
		  value := iter.Value()
		  fmt.Println("key:", string(key))
		  fmt.Println("value:", string(value))
		  fmt.Println("************************************")
	  }
  }else{
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
*/
func (app *SITComApplication) Query(reqQuery types.RequestQuery) (resQuery types.ResponseQuery) {
	switch reqQuery.Path {
	case "/val":
		key := []byte("val:" + string(reqQuery.Data))
		value := app.state.db.Get(key)

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	case "/competencies_from":
		key := []byte("competence_from:" + string(reqQuery.Data))
		value := app.state.db.Get(key)

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	case "/activities_from":
		key := []byte("activities_from:" + string(reqQuery.Data))
		value := app.state.db.Get(key)

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	case "/competencies_semester":
		key := []byte("competencies_semester:" + string(reqQuery.Data))
		value := app.state.db.Get(key)

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	case "/activities_semester":
		key := []byte("activities_semester:" + string(reqQuery.Data))
		value := app.state.db.Get(key)

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	case "/auto_competencies_semester":
		key := []byte("auto_competencies_semester:" + string(reqQuery.Data))
		value := app.state.db.Get(key)

		resQuery.Key = reqQuery.Data
		resQuery.Value = value
		return
	default:
		resQuery.Key = reqQuery.Data
		value := app.state.db.Get(prefixKey(reqQuery.Data))
		resQuery.Value = value
		if value != nil {
			resQuery.Log = "exists"
		} else {
			resQuery.Log = "does not exist"
		}
		return
	}
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
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}

func (app *SITComApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ValidatorSetChangePrefix):]

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

func (app *SITComApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ValidatorSetChangePrefix):]

	pubKeyAndPower := strings.Split(string(tx), "/")
	if len(pubKeyAndPower) != 2 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey/power'. Got %v", pubKeyAndPower)}
	}
	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]

	pubKey, err := hex.DecodeString(pubkeyS)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid hex", pubkeyS)}
	}

	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	return app.updateValidator(types.Ed25519ValidatorUpdate(pubKey, int64(power)))
}

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
	StudentID    	uint64 `json:"student_id"`
	CompetenceID    uint16 `json:"competence_id"`
	By				[]byte `json:"by"`
	Semester		uint16 `json:"semester"`
}

type AttendedActivity struct {
	StudentID 	uint64 `json:"student_id"`
	ActivityID 	uint32 `json:"activity_id"`
	Approver	[]byte `json:"approver"`
	Semester	uint16 `json:"semester"`
}

type AutoAddComptence struct{
	StudentID    	uint64 `json:"student_id"`
	CompetenceID    uint16 `json:"competence_id"`
	Semester		uint16 `json:"semester"`
} 

/*
func (app *SITComApplication) CreateCompetence(args []byte) ([]types.Event, error) {
	var competence Competence
	err := json.Unmarshal(args, &competence)
	if err != nil {
		return nil, err
	}

	key := competence.CompetenceID
	var value []byte

	value = append([]byte(competence.CompetenceName), ';')
	value = append(value, competence.Description...)
	value = append(value, ';')
	value = append(value, strconv.FormatInt(int64(competence.TotalRequiredActivities), 10)...)
	value = append(value, ';')
	value = append(value, competence.Creator...)

	app.state.db.Set([]byte(key), value)
	app.state.Size++

	events := []types.Event{
		{
			Type: "competence.creation",
			Attributes: []cmn.KVPair{
				{Key: []byte("competenceid"), Value: []byte(competence.CompetenceID)},
				{Key: []byte("staffid"), Value: []byte(competence.Creator)},
			},
		},
	}

	return events, nil

}
*/
func (app *SITComApplication) StaffAddCompetence(args []byte) ([]types.Event, error) {
	var update StaffAddCompetence
	err := json.Unmarshal(args, &update)
	if err != nil {
		return nil, err
	}

	value := make([]byte, 20)							// 8 + 2 + 2 + 8 (StudentID + CompetenceID + Semester + state.Size)

	binary.LittleEndian.PutUint64(value, update.StudentID)
	binary.LittleEndian.PutUint16(value, update.CompetenceID)
	value := append(value, update.By...)
	binary.LittleEndian.PutUint16(value, update.Semester)
	binary.LittleEndian.PutUint64(value, app.state.Size + 1)

	key := crypto.Sha256(value)

	app.state.db.Set(key, value)					//Set struct_id to value, without last 8 bytes (state.Size)

	app.state.Size++

	events := []types.Event{
		{
			Type: "competence.add",
			Attributes: []cmn.KVPair{
				{Key: []byte("txid"), Value: key},
				{Key: []byte("studentid"), Value: []byte(strconv.Itoa(update.StudentID))},
				{Key: []byte("competenceid"), Value: []byte(strconv.Itoa(update.StudentID))},
				{Key: []byte("by"), Value: update.By},
				{Key: []byte("semester"), Value: []byte(strconv.Itoa(update.Semester))},
			},
		},
	}

	return events, nil
}

func (app *SITComApplication) AttendedActivity(args []byte) ([]types.Event, error) {
	var update AttendedActivity
	err := json.Unmarshal(args, &update)
	if err != nil {
		return nil, err
	}

	key := append([]byte(update.StudentID), []byte(update.ActivityID)...)
	var value []byte

	value = append([]byte(update.StudentID), ';')
	value = append(value, update.ActivityID...)
	value = append(value, ';')
	value = append(value, update.Approver...)

	app.state.db.Set(key, value)
	app.state.Size++

	events := []types.Event{
		{
			Type: "activity.attend",
			Attributes: []cmn.KVPair{
				{Key: []byte("studentid"), Value: []byte(update.StudentID)},
				{Key: []byte("activityid"), Value: []byte(update.ActivityID)},
				{Key: []byte("staffid"), Value: []byte(update.Approver)},
			},
		},
	}

	return events, nil
}

/*
func (goDb dbm.DB) printAll(){
  str, _ := goDb.db.NewGetProperty("leveldb.stats")
  fmt.Printf("%v\n", str)

  itr := goDb.db.NewIterator(nil,nil)
  for itr.Next(){
    key := itr.Key()
    value := itr.Value()
    fmt.Printf("[%s]:\t[%s]\n",string(key),string(value))
  }
}
*/

