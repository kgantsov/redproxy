package consistent_hashing

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
)

type ConsistentHashing struct {
	nodeMap        map[string]string
	nodes          []string
	nodePartitions []string
	partitions     int
}

func NewConsistentHashing(nodes []string, partitions int) *ConsistentHashing {
	var nodePartitions []string

	nodeMap := make(map[string]string)

	for partition := 0; partition < partitions; partition++ {
		for _, node := range nodes {
			hash := GetMD5Hash(fmt.Sprintf("%d-%s", partition, node))
			nodePartitions = append(nodePartitions, hash)
			nodeMap[hash] = node
		}
	}

	sort.Slice(nodePartitions, func(i, j int) bool {
		return nodePartitions[i] < nodePartitions[j]
	})

	ch := &ConsistentHashing{
		nodes:          nodes,
		partitions:     partitions,
		nodePartitions: nodePartitions,
		nodeMap:        nodeMap,
	}

	return ch
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))

	return hex.EncodeToString(hasher.Sum(nil))
}

func (ch *ConsistentHashing) GetNode(key string) string {
	keyHash := GetMD5Hash(key)

	partitionIndex := sort.Search(
		len(ch.nodePartitions),
		func(i int) bool { return ch.nodePartitions[i] >= keyHash },
	)
	if partitionIndex == len(ch.nodePartitions) {
		partitionIndex = 0
	}

	nodePartition := ch.nodePartitions[partitionIndex]
	node := ch.nodeMap[nodePartition]

	return node
}
