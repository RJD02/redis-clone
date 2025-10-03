package handlers

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// HandleInfo handles the INFO command
func HandleInfo(conn net.Conn, cmd *Command) {
	// Default to all sections if no argument provided
	section := "all"
	if len(cmd.Args) > 0 {
		section = strings.ToLower(cmd.Args[0])
	}

	var infoContent string
	role := config.GetServerRole()

	switch section {
	case "replication":
		// Return replication information with master_replid and master_repl_offset
		if config.IsServerMaster() {
			infoContent = fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%d",
				role, config.Server.MasterReplId, config.Server.MasterReplOffset)
		} else {
			// For slaves, just return the role for now
			infoContent = "role:" + role
		}
	case "all":
		// For now, just return replication info for "all" as well
		// In a full implementation, this would include all sections
		if config.IsServerMaster() {
			infoContent = fmt.Sprintf("# Replication\nrole:%s\nmaster_replid:%s\nmaster_repl_offset:%d",
				role, config.Server.MasterReplId, config.Server.MasterReplOffset)
		} else {
			infoContent = "# Replication\nrole:" + role
		}
	default:
		// Unknown section, return empty (Redis behavior)
		infoContent = ""
	}

	// Encode as bulk string
	response := parser.ToBulkString(infoContent)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write INFO response")
		return
	}

	fmt.Printf("INFO: section=%s, role=%s\n", section, role)
}
