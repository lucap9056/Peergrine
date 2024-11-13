package cluster

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	Storage "peergrine/kafker/storage"

	"github.com/go-zookeeper/zk"
)

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// New initializes a new cluster Manager, setting up the leader election process.
func New(conn *zk.Conn, storage *Storage.Storage) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cluster := &Manager{
		ctx:    ctx,
		cancel: cancel,
	}

	electionPath := "/kafker/leader"

	exists, _, err := conn.Exists(electionPath)
	if err != nil {
		return nil, fmt.Errorf("error checking if path exists: %w", err)
	}

	if !exists {
		_, err = conn.Create(electionPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return nil, fmt.Errorf("error creating election path: %w", err)
		}
	}

	electionNodePath, err := conn.Create(
		electionPath+"/leader_",
		nil,
		zk.FlagEphemeral|zk.FlagSequence,
		zk.WorldACL(zk.PermAll),
	)

	if err != nil {
		return nil, fmt.Errorf("error creating election node: %w", err)
	}

	fmt.Println("Created node:", electionNodePath)

	// Start leader election in a goroutine
	go func() {
		defer func() {
			fmt.Println("Election process exited.")
		}()

		for {
			select {
			case <-cluster.ctx.Done():
				fmt.Println("Shutting down gracefully...")
				return

			default:
				isLeader, err := checkIfLeader(conn, electionPath, electionNodePath)
				if err != nil {
					log.Fatalf("error during leader election: %v", err)
				}

				if isLeader {
					fmt.Println("I am the leader:", electionNodePath)
					storage.StartServiceHealthCheck(cluster.ctx)
				} else {
					fmt.Println("Not the leader, watching for leader changes")
					watchForLeader(conn, electionPath, electionNodePath)
				}

				time.Sleep(2 * time.Second)
			}
		}
	}()

	return cluster, nil
}

// checkIfLeader checks if the current node is the leader by comparing with other nodes.
func checkIfLeader(conn *zk.Conn, electionPath, myNodePath string) (bool, error) {
	nodes, _, err := conn.Children(electionPath)
	if err != nil {
		return false, fmt.Errorf("error fetching children nodes: %w", err)
	}

	sort.Strings(nodes)

	for _, node := range nodes {
		fullPath := electionPath + "/" + node
		if fullPath == myNodePath {
			return true, nil
		} else {
			break
		}
	}

	return false, nil
}

// watchForLeader monitors the previous node for changes.
func watchForLeader(conn *zk.Conn, electionPath, myNodePath string) {
	nodes, _, err := conn.Children(electionPath)
	if err != nil {
		log.Fatalf("error fetching children nodes: %v", err)
	}

	sort.Strings(nodes)

	var prevNode string
	for i, node := range nodes {
		fullPath := electionPath + "/" + node
		if fullPath == myNodePath && i > 0 {
			prevNode = nodes[i-1]
			break
		}
	}

	if prevNode != "" {
		prevNodePath := electionPath + "/" + prevNode
		_, _, ch, err := conn.GetW(prevNodePath)
		if err != nil {
			log.Fatalf("error watching node: %v", err)
		}

		evt := <-ch
		if evt.Type == zk.EventNodeDeleted {
			fmt.Println("Previous leader is gone, checking if I'm the new leader")
		}
	}
}

// Stop gracefully shuts down the cluster manager.
func (c *Manager) Stop() {
	c.cancel()
}
