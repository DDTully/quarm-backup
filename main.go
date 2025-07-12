package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func performBackup(dbUser, dbPass, dbName, dbHost, dbPort, targetDir string) error {
	log.Printf("Starting backup of database '%s'...", dbName)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFileName := fmt.Sprintf("%s_%s.sql", dbName, timestamp)
	backupFilePath := filepath.Join(targetDir, backupFileName)

	cmd := exec.Command("mysqldump",
		"--user="+dbUser,
		"--password="+dbPass,
		"--host="+dbHost,
		"--port="+dbPort,
		"--single-transaction",
		"--routines",
		"--triggers",
		dbName,
	)

	backupFile, err := os.Create(backupFilePath)
	if err != nil {
		return fmt.Errorf("could not create backup file: %w", err)
	}
	defer backupFile.Close()

	cmd.Stdout = backupFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(backupFilePath)
		return fmt.Errorf("failed to execute mysqldump: %w", err)
	}

	log.Printf("✅ Backup completed successfully: %s\n", backupFilePath)
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s -user <user> -pass <pass> -db <db> [options]\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "A command-line tool to back up a MariaDB database using mysqldump.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Required Arguments:")
	fmt.Fprintln(os.Stderr, "  -user string")
	fmt.Fprintln(os.Stderr, "    \tDatabase username")
	fmt.Fprintln(os.Stderr, "  -pass string")
	fmt.Fprintln(os.Stderr, "    \tDatabase password")
	fmt.Fprintln(os.Stderr, "  -db string")
	fmt.Fprintln(os.Stderr, "    \tDatabase name to back up")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Optional Arguments:")
	fmt.Fprintln(os.Stderr, "  -dir string")
	fmt.Fprintln(os.Stderr, "    \tTarget directory for the backup file (default \".\")")
	fmt.Fprintln(os.Stderr, "  -host string")
	fmt.Fprintln(os.Stderr, "    \tDatabase host (default \"127.0.0.1\")")
	fmt.Fprintln(os.Stderr, "  -port string")
	fmt.Fprintln(os.Stderr, "    \tDatabase port (default \"3306\")")
	fmt.Fprintln(os.Stderr, "  -t int")
	fmt.Fprintln(os.Stderr, "    \tTime in minutes to repeat the backup. If 0, runs only once.")
	fmt.Fprintln(os.Stderr)
}

func main() {
	dbUser := flag.String("user", "", "Database username")
	dbPass := flag.String("pass", "", "Database password")
	dbName := flag.String("db", "", "Database name to back up")
	targetDir := flag.String("dir", ".", "Target directory for the backup file")
	dbHost := flag.String("host", "127.0.0.1", "Database host")
	dbPort := flag.String("port", "3306", "Database port")
	repeatMinutes := flag.Int("t", 0, "Cadence in minutes to repeat the backup")

	flag.Usage = printUsage
	flag.Parse()

	if *dbUser == "" || *dbPass == "" || *dbName == "" {
		flag.Usage()
		os.Exit(1)
	}

	_, err := exec.LookPath("mysqldump")
	if err != nil {
		log.Fatalf("Fatal Error: 'mysqldump' command not found in PATH. Please install MariaDB/MySQL client tools.")
	}

	dirInfo, err := os.Stat(*targetDir)
	if os.IsNotExist(err) || !dirInfo.IsDir() {
		log.Fatalf("Fatal Error: Target directory '%s' does not exist or is not a directory.", *targetDir)
	}

	if *repeatMinutes <= 0 {
		log.Println("Performing a single backup.")
		err := performBackup(*dbUser, *dbPass, *dbName, *dbHost, *dbPort, *targetDir)
		if err != nil {
			log.Fatalf("Fatal Error: Backup failed: %v", err)
		}
	} else {
		log.Printf("Starting backup daemon to run every %d minute(s).", *repeatMinutes)
		log.Println("Performing initial backup now...")

		if err := performBackup(*dbUser, *dbPass, *dbName, *dbHost, *dbPort, *targetDir); err != nil {
			log.Printf("ERROR: Initial backup failed: %v", err)
		}

		ticker := time.NewTicker(time.Duration(*repeatMinutes) * time.Minute)
		defer ticker.Stop()

		log.Printf("Waiting %d minute(s). Press Ctrl+C to exit.", *repeatMinutes)

		for range ticker.C {
			if err := performBackup(*dbUser, *dbPass, *dbName, *dbHost, *dbPort, *targetDir); err != nil {
				log.Printf("ERROR: Scheduled backup failed: %v", err)
			} else {
				log.Printf("Waiting %d minute(s). Press Ctrl+C to exit.", *repeatMinutes)
			}
		}
	}
}
