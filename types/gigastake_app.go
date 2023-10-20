package types

import (
	"time"
)

/* GigastakeApp Struct Definition and Methods */
type (
	// GigastakeApp represents a single gigastaked application for a given chain
	GigastakeApp struct {
		ID              GigastakeAppID            `json:"id"`
		ChainIDs        map[RelayChainID]struct{} `json:"chainIDs"`
		Name            string                    `json:"name"`
		Address         string                    `json:"address"`
		PublicKey       GigastakeAppPublicKey     `json:"publicKey"`
		ClientPublicKey string                    `json:"clientPublicKey"`
		Signature       string                    `json:"signature"`
		Version         string                    `json:"version"`
		CreatedAt       time.Time                 `json:"createdAt"`
		UpdatedAt       time.Time                 `json:"updatedAt"`
		Deleted         bool                      `json:"deleted"`
	}

	// UpdateGigastakeApp represents the fields that can be updated on a GigastakeApp
	UpdateGigastakeApp struct {
		ID       GigastakeAppID `json:"id"`
		Name     string         `json:"name"`
		ChainIDs []RelayChainID `json:"chainIDs"`
	}

	// GigastakeApp represents the relationship between a Chain and GigastakeApp
	// Only used by the listener
	ChainGigastakeApp struct {
		ChainID        RelayChainID   `json:"chainID"`
		GigastakeAppID GigastakeAppID `json:"gigastakeAppID"`
	}

	GigastakeAppPublicKey string
)

func (a *GigastakeApp) Table() Table {
	return TableGigastakeApps
}

func (a *ChainGigastakeApp) Table() Table {
	return TableChainGigastakeApps
}
