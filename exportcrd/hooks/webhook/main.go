package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// APILifecycleWebhookRequest represents the webhook HTTP request payload format.
// TODO: consider using CloudEvents
// +kubebuilder:object:generate=false
type APILifecycleWebhookRequest struct {
	// WorkspaceID is the identifier of the workspace containing the APIBinding.
	WorkspaceID string `json:"workspaceID"` // TODO: consider naming it `source` (CloudEvents)

	// Type is the fully-qualified workspace type name
	WorkspaceTypeName string `json:"workspaceTypeName"`

	// APIBindingUID is the APIBinding UID
	APIBindingUID string `json:"apiBindingUID"` // TODO: consider naming it `id` (CloudEvents)

	// Type is the type of event that triggered the webhook.
	Type string `json:"type"`
}

func handle(w http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)

	if err != nil {
		fmt.Printf("failed to read the request body: %v\n", err)
		w.WriteHeader(500)
		return
	}

	var body APILifecycleWebhookRequest
	err = json.Unmarshal(b, &body)
	if err != nil {
		fmt.Printf("failed to unmarshal the request body: %v\n", err)
		w.WriteHeader(500)
		return
	}

	fmt.Printf("workspace %s, of type %s, bound via apibinding %s\n", body.WorkspaceID, body.WorkspaceTypeName, body.APIBindingUID)

	manifests := `[
			{
				"resource": "secrets",
				"content": {
					"apiVersion": "v1",
          "kind": "Secret",
					"metadata": {
						"name": "from-provider"
					},
					"stringData": {
						"avery": "secret"
					}
        }
			}
		]`

	w.Write([]byte(manifests))
}

func main() {
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
