package app

import (
	"encoding/json"

	"github.com/saguywalker/sitcomchain/code"
	"github.com/tendermint/tendermint/abci/types"
)

func (a *SitcomApplication) addNewService(payload []byte) (res types.ResponseDeliverTx, err error) {
	return
}

func (a *SitcomApplication) giveBadge(payload []byte) (res types.ResponseDeliverTx, err error) {
	var sorted map[string]interface{}
	if err := json.Unmarshal(payload, &sorted); err != nil {
		res.Code = code.CodeTypeUnmarshalError
		res.Log = "error when unmarshal params"
		return res, err
	}

	delete(sorted, "giver")

	badgeKey, err := json.Marshal(sorted)
	if err != nil {
		res.Code = code.CodeTypeEncodingError
		res.Log = "error when marshal badgeKey"
		return res, err
	}

	a.logger.Infof("k: %s, v: %s\n", badgeKey, payload)
	a.state.db.Set(badgeKey, payload)
	a.state.Size++
	res.Code = code.CodeTypeOK
	res.Log = "success"

	return res, nil
}

func (a *SitcomApplication) approveActivity(payload []byte) (res types.ResponseDeliverTx, err error) {
	var sorted map[string]interface{}
	if err := json.Unmarshal(payload, &sorted); err != nil {
		res.Code = code.CodeTypeUnmarshalError
		res.Log = "error when unmarshal params"
		return res, err
	}

	delete(sorted, "approver")

	activityKey, err := json.Marshal(sorted)
	if err != nil {
		res.Code = code.CodeTypeEncodingError
		res.Log = "error when marshal badgeKey"
		return res, err
	}

	a.logger.Infof("k: %s, v: %s\n", activityKey, payload)
	a.state.db.Set(activityKey, payload)
	a.state.Size++
	res.Code = code.CodeTypeOK
	res.Log = "success"

	return res, nil
}
