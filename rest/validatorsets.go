package rest

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/jim380/Cendermint/config"
	"go.uber.org/zap"
)

type validatorsetsLegacy struct {
	Height string `json:"height"`

	Result struct {
		Block_Height string `json:"block_height"`
		Validators   []struct {
			ConsAddr         string                 `json:"address"`
			ConsPubKey       consPubKeyValSetLegacy `json:"pub_key"`
			ProposerPriority string                 `json:"proposer_priority"`
			VotingPower      string                 `json:"voting_power"`
		}
	}
}

type validatorsets struct {
	Block_Height string `json:"block_height"`
	Validators   []struct {
		ConsAddr         string           `json:"address"`
		ConsPubKey       consPubKeyValSet `json:"pub_key"`
		ProposerPriority string           `json:"proposer_priority"`
		VotingPower      string           `json:"voting_power"`
	} `json:"validators"`
}

type consPubKeyValSetLegacy struct {
	Type string `json:"type"`
	Key  string `json:"value"`
}

type consPubKeyValSet struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

func (rd *RESTData) getValidatorsets(cfg config.Config, currentBlockHeight int64) []string {
	var vSetsResultFinal map[string][]string

	if cfg.IsLegacySDKVersion() {
		var vSets, vSets2, vSets3, vsetTest validatorsetsLegacy
		var vSetsResult map[string][]string = make(map[string][]string)
		var vSetsResult2 map[string][]string = make(map[string][]string)
		var vSetsResult3 map[string][]string = make(map[string][]string)

		shouldRunPages := testPageLimit(cfg, currentBlockHeight, &vsetTest, 3)

		if shouldRunPages {
			runPages(cfg, currentBlockHeight, &vSets, vSetsResult, 1)
			runPages(cfg, currentBlockHeight, &vSets2, vSetsResult2, 2)
			runPages(cfg, currentBlockHeight, &vSets3, vSetsResult3, 3)

			for _, value := range vSets.Result.Validators {
				// populate the validatorset map => [ConsPubKey][]string{ConsAddr, VotingPower, ProposerPriority}
				vSetsResult[value.ConsPubKey.Key] = []string{value.ConsAddr, value.VotingPower, value.ProposerPriority, "0"}
			}

			for _, value := range vSets2.Result.Validators {
				// populate the validatorset map => [ConsPubKey][]string{ConsAddr, VotingPower, ProposerPriority}
				vSetsResult2[value.ConsPubKey.Key] = []string{value.ConsAddr, value.VotingPower, value.ProposerPriority, "0"}
			}

			for _, value := range vSets3.Result.Validators {
				// populate the validatorset map => [ConsPubKey][]string{ConsAddr, VotingPower, ProposerPriority}
				vSetsResult3[value.ConsPubKey.Key] = []string{value.ConsAddr, value.VotingPower, value.ProposerPriority, "0"}
			}
			vSetsResultTemp := mergeMap(vSetsResult, vSetsResult2)
			vSetsResultFinal = mergeMap(vSetsResultTemp, vSetsResult3)
			zap.L().Info("", zap.Bool("Success", true), zap.String("Active validators", fmt.Sprint(len(vSets.Result.Validators)+len(vSets2.Result.Validators)+len(vSets3.Result.Validators))))
		} else {
			runPages(cfg, currentBlockHeight, &vSets, vSetsResult, 1)
			for _, value := range vSets.Result.Validators {
				// populate the validatorset map => [ConsPubKey][]string{ConsAddr, VotingPower, ProposerPriority}
				vSetsResult[value.ConsPubKey.Key] = []string{value.ConsAddr, value.VotingPower, value.ProposerPriority, "0"}
			}
			vSetsResultFinal = vSetsResult
		}
	} else {
		var vSets validatorsets
		var vSetsResult map[string][]string = make(map[string][]string)

		route := getValidatorSetByHeightRoute(cfg)
		res, err := HttpQuery(RESTAddr + route + fmt.Sprint(currentBlockHeight))
		if err != nil {
			zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", err.Error()))
		}

		json.Unmarshal(res, &vSets)

		if strings.Contains(string(res), "not found") {
			zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", string(res)))
		} else if strings.Contains(string(res), "error:") || strings.Contains(string(res), "error\\\":") {
			zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", string(res)))
		}

		for _, value := range vSets.Validators {
			// populate the validatorset map => [ConsPubKey][]string{ConsAddr, VotingPower, ProposerPriority}
			vSetsResult[value.ConsPubKey.Key] = []string{value.ConsAddr, value.VotingPower, value.ProposerPriority, "0"}
		}

		vSetsResultFinal = vSetsResult
		zap.L().Info("", zap.Bool("Success", true), zap.String("Active validators", fmt.Sprint(len(vSets.Validators))))
	}

	rd.Validatorsets = Sort(vSetsResultFinal, 2) // sort by ProposerPriority
	for key, value := range rd.Validatorsets {
		zap.L().Debug("", zap.Bool("Success", true), zap.String(key, strings.Join(value, ", ")))
	}

	if len(rd.Validatorsets) == 0 {
		zap.L().Warn("", zap.Bool("Success", false), zap.String("err", "Validator set is empty"))
	}

	rd.getValidator(cfg)
	valInfo := rd.locateValidatorInActiveSet()
	return valInfo
}

// TO-DO if consumer chain, use cosmoshub's ConsPubKey
func (rd *RESTData) locateValidatorInActiveSet() []string {
	valInfo, found := rd.Validatorsets[rd.Validator.ConsPubKey.Key]
	if !found {
		zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", "Validator not found in the active set"))
	}
	return valInfo
}

func Sort(mapValue map[string][]string, index int) map[string][]string {
	keys := make([]string, 0, len(mapValue))
	for k := range mapValue {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		vi, _ := strconv.Atoi(mapValue[keys[i]][1])
		vj, _ := strconv.Atoi(mapValue[keys[j]][1])
		return vi > vj
	})

	sortedVSetsResult := make(map[string][]string)
	for _, k := range keys {
		sortedVSetsResult[k] = mapValue[k]
	}

	return sortedVSetsResult
}

func mergeMap(a map[string][]string, b map[string][]string) map[string][]string {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func runPages(cfg config.Config, currentBlockHeight int64, vSets *validatorsetsLegacy, vSetsResult map[string][]string, pages int) {
	route := getValidatorSetByHeightRoute(cfg)

	res, err := HttpQuery(RESTAddr + route + fmt.Sprint(currentBlockHeight))

	if err != nil {
		zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", err.Error()))
	}

	json.Unmarshal(res, &vSets)

	if strings.Contains(string(res), "not found") {
		zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", string(res)))
	} else if strings.Contains(string(res), "error:") || strings.Contains(string(res), "error\\\":") {
		zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", string(res)))
	}

	for _, value := range vSets.Result.Validators {
		// populate the validatorset map => [ConsPubKey][]string{ConsAddr, VotingPower, ProposerPriority}
		vSetsResult[value.ConsPubKey.Key] = []string{value.ConsAddr, value.VotingPower, value.ProposerPriority, "0"}
	}
}

func testPageLimit(cfg config.Config, currentBlockHeight int64, vSets *validatorsetsLegacy, maxPageNumber int64) bool {
	multiPagesSupported := true

	route := getValidatorSetByHeightRoute(cfg)
	res, err := HttpQuery(RESTAddr + route + fmt.Sprint(currentBlockHeight) + "?page=3")
	if err != nil {
		zap.L().Fatal("", zap.Bool("Success", false), zap.String("err", err.Error()))
	}

	json.Unmarshal(res, &vSets)

	if strings.Contains(string(res), "Internal error: page should be within") {
		zap.L().Info("", zap.String("warn", string(res)))
		multiPagesSupported = false
	}

	return multiPagesSupported
}
