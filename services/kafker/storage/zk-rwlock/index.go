package zkrwlock

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	ZKRLOCK = "rlock-"
	ZKWLOCK = "wlock-"
)

type ZkWLock struct {
	conn     *zk.Conn
	nodePath string
}

func WLock(conn *zk.Conn, path string) (*ZkWLock, error) {
	lockPrefix := path + "/" + ZKWLOCK
	log.Printf("Attempting to create write lock at %s", lockPrefix)

	nodePath, err := conn.CreateProtectedEphemeralSequential(lockPrefix, nil, zk.WorldACL(zk.PermAll))
	if err != nil {
		return nil, fmt.Errorf("failed to create write lock: %w", err)
	}
	log.Printf("Created write lock node %s", nodePath)

	initialDelay := time.Second * 2
	maxDelay := time.Second * 10
	delay := initialDelay

	for {
		nodes, _, eventChan, err := conn.ChildrenW(path)
		if err != nil {
			conn.Delete(nodePath, -1)
			return nil, fmt.Errorf("failed to get children of %s: %w", path, err)
		}
		log.Printf("Retrieved children for path %s: %v", path, nodes)

		if !isWriteLocked(nodes, nodePath) {
			log.Printf("Acquired write lock at node %s", nodePath)
			return &ZkWLock{conn: conn, nodePath: nodePath}, nil
		}

		log.Printf("Write lock not acquired, waiting on lock. Current delay: %v", delay)
		if err := waitForLock(eventChan, &delay, initialDelay, maxDelay); err != nil {
			conn.Delete(nodePath, -1)
			return nil, err
		}
	}
}

// WUnlock releases the write lock.
func (l *ZkWLock) WUnlock() error {
	log.Printf("Releasing write lock at node %s", l.nodePath)
	if err := l.conn.Delete(l.nodePath, -1); err != nil {
		log.Printf("Failed to release write lock: %s, error: %v", l.nodePath, err)
		return err
	}
	log.Printf("Write lock released at node %s", l.nodePath)
	return nil
}

type ZkRLock struct {
	conn     *zk.Conn
	nodePath string
}

// RLock acquires a read lock.
func RLock(conn *zk.Conn, path string) (*ZkRLock, error) {
	initialDelay := time.Second * 2
	maxDelay := time.Second * 10
	delay := initialDelay

	for {
		nodes, _, eventChan, err := conn.ChildrenW(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get children of %s: %w", path, err)
		}
		log.Printf("Retrieved children for path %s: %v", path, nodes)

		sort.Strings(nodes)

		if !isReadLocked(nodes) {
			rLockPrefix := path + "/" + ZKRLOCK
			log.Printf("Attempting to create read lock at %s", rLockPrefix)
			nodePath, err := conn.CreateProtectedEphemeralSequential(rLockPrefix, nil, zk.WorldACL(zk.PermAll))
			if err != nil {
				return nil, fmt.Errorf("failed to create read lock: %w", err)
			}
			log.Printf("Acquired read lock at node %s", nodePath)
			return &ZkRLock{conn: conn, nodePath: nodePath}, nil
		}

		log.Printf("Read lock not acquired, waiting on lock. Current delay: %v", delay)
		if err := waitForLock(eventChan, &delay, initialDelay, maxDelay); err != nil {
			return nil, err
		}
	}
}

// RUnlock releases the read lock.
func (l *ZkRLock) RUnlock() error {
	log.Printf("Releasing read lock at node %s", l.nodePath)
	if err := l.conn.Delete(l.nodePath, -1); err != nil {
		return fmt.Errorf("failed to release read lock %s: %w", l.nodePath, err)
	}
	log.Printf("Read lock released at node %s", l.nodePath)
	return nil
}

func isWriteLocked(nodes []string, currentPath string) bool {
	sort.Strings(nodes)
	wIndex := 0
	for _, node := range nodes {
		if strings.Contains(node, ZKRLOCK) {
			return true
		} else if strings.Contains(node, ZKWLOCK) {

			if strings.Contains(currentPath, node) && wIndex == 0 {
				return false
			} else {
				wIndex++
			}

		}
	}

	return true
}

func isReadLocked(nodes []string) bool {
	for _, node := range nodes {
		if strings.Contains(node, ZKWLOCK) {
			return true
		}
	}

	return false
}

// waitForLock handles waiting for the lock and increasing delay.
func waitForLock(eventChan <-chan zk.Event, delay *time.Duration, initialDelay, maxDelay time.Duration) error {
	select {
	case <-eventChan:
		log.Printf("Event received, resetting delay to initial %v", initialDelay)
		*delay = initialDelay
		return nil
	case <-time.After(*delay):
		if *delay < maxDelay {
			*delay *= 2
			if *delay > maxDelay {
				*delay = maxDelay
			}
			log.Printf("No event received, increasing delay to %v", *delay)
		}
	}
	return nil
}
