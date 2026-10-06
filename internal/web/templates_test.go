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

func TestFirewallRequestsListRendersRuleDescription(t *testing.T) {
	file := filepath.Join("..", "..", "www", "html", "js", "firewall-requests.js")
	body, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile(%q) failed: %v", file, err)
	}

	content := string(body)
	if !strings.Contains(content, "<td>${esc(item.rule_description)}</td>") {
		t.Fatalf("firewall request list row does not render rule_description")
	}
	if strings.Contains(content, "<td>${esc(item.firewall_zone)}<br><span class=\"hint mono\">${esc(item.firewall_id)}</span></td>") {
		t.Fatalf("firewall request list row still renders firewall zone and ID in the rule description column")
	}
}

func TestFirewallRequestActiveStatusDisplaysInUse(t *testing.T) {
	files := []string{
		filepath.Join("..", "..", "www", "html", "js", "firewall-requests.js"),
		filepath.Join("..", "..", "www", "html", "partials", "firewall-request-modal.html"),
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			body, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("ReadFile(%q) failed: %v", file, err)
			}

			content := string(body)
			if !strings.Contains(content, "使用中") {
				t.Fatalf("firewall request active status label should contain 使用中")
			}
		})
	}
}
