package web

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirewallRequestsHeaderUsesRuleDescription(t *testing.T) {
	files := []string{
		filepath.Join("..", "..", "www", "template", "firewall_requests.gohtml"),
		filepath.Join("..", "..", "www", "html", "firewall-requests.html"),
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("ReadFile(%q) failed: %v", file, err)
			}

			content := string(body)
			if !strings.Contains(content, "<th>規則說明</th>") {
				t.Fatalf("firewall request table header = missing 規則說明")
			}
			if strings.Contains(content, "<th>Firewall 區域 / ID</th>") {
				t.Fatalf("firewall request table header still contains old Firewall 區域 / ID label")
			}
		})
	}
}
