package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var excludeFilePaths = []string{
	".git",
	"scripts",
	".dockerignore",
	".env",
	".gitignore",
	"docker-compose-development.yml",
	"docker-compose-staging.yml",
	"LICENSE",
	"mailserver_setup.sh",
	"mailserver.env",
	"docker-data",
}

func main() {
	projectRootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	projectRootDir = fmt.Sprintf("%s/..", projectRootDir) // Navigate one folder above the script path

	sshConfigIdentify := "~/.ssh/<somekey>"    // SSH configuration options (e.g., "-i /path/to/ssh_key")
	remoteUser := "your_username"              // Your username on the remote server
	remoteHost := "your_server_ip_or_hostname" // The IP or hostname of the remote server
	remoteProjectDir := "/saas"                // Remote path for the project directory

	fmt.Println("Copying the entire project directory to the server...")
	if err := copyProjectToServer(projectRootDir, remoteUser, remoteHost, remoteProjectDir, sshConfigIdentify); err != nil {
		log.Fatalf("Failed to copy files to the server: %v", err)
	}

	fmt.Println("Logging in to the server to rebuild and restart services...")
	if err := rebuildAndRestartServices(remoteUser, remoteHost, remoteProjectDir, sshConfigIdentify); err != nil {
		log.Fatalf("Failed to rebuild and restart services: %v", err)
	}
}

func isExcludedFileOrFolder(path string) bool {
	for _, excluded := range excludeFilePaths {
		if strings.Contains(path, excluded) {
			return true
		}
	}
	return false
}

func copyProjectToServer(projectRootDir, remoteUser, remoteHost, remoteProjectDir, sshConfigIdentify string) error {
	client, err := connectSFTP(remoteUser, remoteHost, sshConfigIdentify)
	if err != nil {
		return err
	}
	defer client.Close()

	walker, err := client.WalkDir(projectRootDir)
	if err != nil {
		return fmt.Errorf("error while walking through the directory: %w", err)
	}

	for walker.Step() {
		if err := walker.Err(); err != nil {
			return fmt.Errorf("error while walking through the directory: %w", err)
		}
		localPath := walker.Path()
		remotePath := strings.Replace(localPath, projectRootDir, remoteProjectDir, 1)

		if isExcludedFileOrFolder(localPath) {
			continue
		}

		info := walker.Info()
		if info.IsDir() {
			err := client.MkdirAll(remotePath)
			if err != nil {
				return fmt.Errorf("failed to create remote directory: %w", err)
			}
		} else {
			localFile, err := os.Open(localPath)
			if err != nil {
				return fmt.Errorf("failed to open local file: %w", err)
			}
			defer localFile.Close()

			remoteFile, err := client.Create(remotePath)
			if err != nil {
				return fmt.Errorf("failed to create remote file: %w", err)
			}
			defer remoteFile.Close()

			_, err = io.Copy(remoteFile, localFile)
			if err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
		}
	}
	return nil
}

func loadLocalEnvToServer(remoteUser, remoteHost, remoteProjectDir, sshConfigIdentify string) error {
	envFilePath := fmt.Sprintf("%s/.env", remoteProjectDir)
	envFile, err := os.Open(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local .env file: %w", err)
	}
	defer envFile.Close()

	var envVars []string
	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		envVars = append(envVars, fmt.Sprintf("export %s", line))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read local .env file: %w", err)
	}

	client, session, err := connectSSH(remoteUser, remoteHost, sshConfigIdentify)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	cmd := strings.Join(envVars, "\n")
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to set environment variables on server: %w", err)
	}

	fmt.Println("Environment variables loaded successfully.")
	return nil
}

func rebuildAndRestartServices(remoteUser, remoteHost, remoteProjectDir, sshConfigIdentify string) error {
	client, session, err := connectSSH(remoteUser, remoteHost, sshConfigIdentify)
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	cmd := fmt.Sprintf("cd %s && docker-compose -f docker-compose.production.yml up -d --build backend frontend", remoteProjectDir)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	fmt.Println("Services rebuilt and restarted successfully.")
	return nil
}

func connectSFTP(remoteUser, remoteHost, sshConfigIdentify string) (*sftp.Client, error) {
	client, err := connectSSHClient(remoteUser, remoteHost, sshConfigIdentify)
	if err != nil {
		return nil, err
	}
	defer client.Close() // Ensure client is closed in case of an error

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	return sftpClient, nil
}

func connectSSH(remoteUser, remoteHost, sshConfigIdentify string) (*ssh.Client, *ssh.Session, error) {
	client, err := connectSSHClient(remoteUser, remoteHost, sshConfigIdentify)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	return client, session, nil
}

func connectSSHClient(remoteUser, remoteHost, sshConfigIdentify string) (*ssh.Client, error) {
	authMethod, err := loadSSHKey(sshConfigIdentify)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key: %w", err)
	}

	clientConfig := &ssh.ClientConfig{
		User:            remoteUser,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", remoteHost+":22", clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}
	return client, nil
}

func loadSSHKey(path string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}
