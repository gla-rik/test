//go:build wireinject
// +build wireinject

//go:generate wire

package dependency

import (
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(ProviderSet)
	return nil, nil
}
