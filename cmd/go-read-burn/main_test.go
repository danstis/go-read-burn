package main

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCreateDBDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case: Directory does not exist
	dbPath := tempDir + "/db/secrets.db"
	err = createDBDir(dbPath)
	if err != nil {
		t.Errorf("Failed to create directory: %v", err)
	}

	// Check if directory was created
	_, err = os.Stat(tempDir + "/db")
	if os.IsNotExist(err) {
		t.Errorf("Directory was not created")
	}

	// Test case: Directory already exists
	err = createDBDir(dbPath)
	if err != nil {
		t.Errorf("Failed when directory already exists: %v", err)
	}
}

func TestIndexHandler(t *testing.T) {
	// Initialize templates
	templates = template.Must(template.ParseFS(views, "views/*.html"))

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(IndexHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Execute template to get expected HTML content
	var index bytes.Buffer
	if err := templates.ExecuteTemplate(&index, "index.html", nil); err != nil {
		t.Fatal(err)
	}

	expected := index.String()
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestSecretTemplateXSSProtection(t *testing.T) {
	// Initialize templates
	templates = template.Must(template.ParseFS(views, "views/*.html"))

	tests := []struct {
		name           string
		secret         string
		shouldContain  string
		shouldNotMatch string
	}{
		{
			name:           "script tag XSS payload",
			secret:         "<script>alert(1)</script>",
			shouldContain:  "&lt;script&gt;alert(1)&lt;/script&gt;",
			shouldNotMatch: "<script>alert(1)</script>",
		},
		{
			name:           "img onerror XSS payload",
			secret:         "<img onerror=alert(1)>",
			shouldContain:  "&lt;img onerror=alert(1)&gt;",
			shouldNotMatch: "<img onerror=alert(1)>",
		},
		{
			name:           "event handler XSS payload",
			secret:         "onclick=\"alert('xss')\"",
			shouldContain:  "onclick=&#34;alert(&#39;xss&#39;)&#34;",
			shouldNotMatch: "onclick=\"alert('xss')\"",
		},
		{
			name:           "svg onclick XSS payload",
			secret:         "<svg onload=alert('xss')></svg>",
			shouldContain:  "&lt;svg onload=alert(&#39;xss&#39;)&gt;&lt;/svg&gt;",
			shouldNotMatch: "<svg onload=alert('xss')></svg>",
		},
		{
			name:           "normal secret is preserved",
			secret:         "This is a safe secret message",
			shouldContain:  "This is a safe secret message",
			shouldNotMatch: "____NO_MATCH____",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Render the secret.html template with the XSS payload
			var output bytes.Buffer
			data := map[string]interface{}{
				"Secret": tt.secret,
			}

			if err := templates.ExecuteTemplate(&output, "secret.html", data); err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			result := output.String()

			// Verify that HTML entities are present (escaped)
			if !bytes.Contains(output.Bytes(), []byte(tt.shouldContain)) {
				t.Errorf("Expected escaped content %q not found in output", tt.shouldContain)
				t.Logf("Output snippet: %s", result[0:min(len(result), 500)])
			}

			// Verify that raw XSS payload is NOT present in the output
			if bytes.Contains(output.Bytes(), []byte(tt.shouldNotMatch)) {
				t.Errorf("Raw XSS payload %q should not be present in output", tt.shouldNotMatch)
				t.Logf("Output snippet: %s", result[0:min(len(result), 500)])
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
