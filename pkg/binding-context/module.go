package bindingcontext

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetBindingContexts() ([]BindingContext, error) {
	contexts := make([]BindingContext, 0)
	contextsContent, err := os.Open(os.Getenv("BINDING_CONTEXT_PATH"))
	if err != nil {
		return nil, fmt.Errorf("open binding context file: %w", err)
	}

	err = json.NewDecoder(contextsContent).Decode(&contexts)
	if err != nil {
		return nil, fmt.Errorf("decode binding context: %w", err)
	}

	return contexts, nil
}
