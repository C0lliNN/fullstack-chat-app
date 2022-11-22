package generator

import "github.com/google/uuid"

type CodeGenerator struct {}

// NewCode that's a simple implementation that would need collision check upstream
func (g CodeGenerator) NewCode() string {
	return uuid.NewString()[:6]
}