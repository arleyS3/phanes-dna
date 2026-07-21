package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFile_JavaSample(t *testing.T) {
	tmpDir := t.TempDir()
	samplePath := filepath.Join(tmpDir, "UserController.java")

	sampleCode := `package com.example.demo.controller;

import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.bind.annotation.GetMapping;
import com.example.demo.service.UserService;

@RestController
public class UserController {
    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping("/users")
    public String getUsers() {
        return "users";
    }
}
`
	if err := os.WriteFile(samplePath, []byte(sampleCode), 0644); err != nil {
		t.Fatalf("failed to write sample file: %v", err)
	}

	sf, err := AnalyzeFile(samplePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	if sf.Language != "java" {
		t.Errorf("expected language 'java', got %q", sf.Language)
	}

	// Verify layer classification detects Controller
	foundController := false
	for _, layer := range sf.Layers {
		if layer == "controller" {
			foundController = true
			break
		}
	}

	if !foundController {
		t.Errorf("expected layer 'controller', got %v", sf.Layers)
	}

	if len(sf.AST) == 0 {
		t.Errorf("expected AST nodes to be extracted, got 0")
	}
}
