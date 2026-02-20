package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func sortConfigs() {
	fmt.Println("Starting protocol-based config sorting...")
	
	// Setup paths for new directory structure in current directory
	protocolDir := "Splitted-By-Protocol"
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(protocolDir, 0755); err != nil {
		fmt.Printf("Error creating protocol directory: %v\n", err)
		return
	}
	
	// Define file paths
	files := map[string]string{
		"vmess":  filepath.Join(protocolDir, "vmess.txt"),
		"vless":  filepath.Join(protocolDir, "vless.txt"),
		"trojan": filepath.Join(protocolDir, "trojan.txt"),
		"ss":     filepath.Join(protocolDir, "ss.txt"),
		"ssr":    filepath.Join(protocolDir, "ssr.txt"),
	}
	
	// Clear existing files
	for protocol, filePath := range files {
		if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
			fmt.Printf("Error clearing %s file: %v\n", protocol, err)
			return
		}
	}
	
	// Fetch the main config file
	fmt.Println("Fetching main configuration file...")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", "https://raw.githubusercontent.com/Epodonios/v2ray-configs/main/All_Configs_Sub.txt", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching config file: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP error: %d\n", resp.StatusCode)
		return
	}
	
	// Process the response line by line for memory efficiency
	scanner := bufio.NewScanner(resp.Body)
	
	// Collect configs by protocol
	protocolConfigs := make(map[string][]string)
	// Track duplicates for each protocol
	seenConfigs := make(map[string]map[string]bool)
	for protocol := range files {
		seenConfigs[protocol] = make(map[string]bool)
	}
	
	vmessFile, err := os.OpenFile(files["vmess"], os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening vmess file: %v\n", err)
		return
	}
	defer vmessFile.Close()
	
	vmessWriter := bufio.NewWriter(vmessFile)
	defer vmessWriter.Flush()
	
	configCount := make(map[string]int)
	duplicateCount := make(map[string]int)
	
	fmt.Println("Processing configurations...")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		// Check protocol and categorize
		switch {
		case strings.HasPrefix(line, "vmess://"):
			// Check for duplicates in vmess
			if seenConfigs["vmess"][line] {
				duplicateCount["vmess"]++
				continue
			}
			seenConfigs["vmess"][line] = true
			// Write vmess configs directly to file to save memory
			if _, err := vmessWriter.WriteString(line + "\n"); err != nil {
				fmt.Printf("Error writing vmess config: %v\n", err)
				return
			}
			configCount["vmess"]++
			
		case strings.HasPrefix(line, "vless://"):
			// Check for duplicates in vless
			if seenConfigs["vless"][line] {
				duplicateCount["vless"]++
				continue
			}
			seenConfigs["vless"][line] = true
			protocolConfigs["vless"] = append(protocolConfigs["vless"], line)
			configCount["vless"]++
			
		case strings.HasPrefix(line, "trojan://"):
			// Check for duplicates in trojan
			if seenConfigs["trojan"][line] {
				duplicateCount["trojan"]++
				continue
			}
			seenConfigs["trojan"][line] = true
			protocolConfigs["trojan"] = append(protocolConfigs["trojan"], line)
			configCount["trojan"]++
			
		case strings.HasPrefix(line, "ss://"):
			// Check for duplicates in ss
			if seenConfigs["ss"][line] {
				duplicateCount["ss"]++
				continue
			}
			seenConfigs["ss"][line] = true
			protocolConfigs["ss"] = append(protocolConfigs["ss"], line)
			configCount["ss"]++
			
		case strings.HasPrefix(line, "ssr://"):
			// Check for duplicates in ssr
			if seenConfigs["ssr"][line] {
				duplicateCount["ssr"]++
				continue
			}
			seenConfigs["ssr"][line] = true
			protocolConfigs["ssr"] = append(protocolConfigs["ssr"], line)
			configCount["ssr"]++
		}
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}
	
	// Flush vmess writer
	vmessWriter.Flush()
	
	// Write other protocols as base64-encoded content
	for protocol, configs := range protocolConfigs {
		if len(configs) == 0 {
			continue
		}
		
		// Join all configs for this protocol
		content := strings.Join(configs, "\n")
		
		// Base64 encode the content
		encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
		
		// Write to file
		if err := os.WriteFile(files[protocol], []byte(encodedContent), 0644); err != nil {
			fmt.Printf("Error writing %s file: %v\n", protocol, err)
			return
		}
	}
	
	// Print summary
	fmt.Println("\nProtocol sorting completed!")
	fmt.Println("Configuration counts (after removing duplicates):")
	for protocol, count := range configCount {
		fmt.Printf("  %s: %d configs\n", protocol, count)
	}
	
	total := 0
	totalDuplicates := 0
	for _, count := range configCount {
		total += count
	}
	for _, count := range duplicateCount {
		totalDuplicates += count
	}
	fmt.Printf("  Total unique: %d configs\n", total)
	
	if totalDuplicates > 0 {
		fmt.Println("\nDuplicates removed:")
		for protocol, count := range duplicateCount {
			if count > 0 {
				fmt.Printf("  %s: %d duplicates\n", protocol, count)
			}
		}
		fmt.Printf("  Total duplicates removed: %d\n", totalDuplicates)
		fmt.Printf("  Original total: %d configs\n", total+totalDuplicates)
	}
}
