package main

import (
    "sort"
	"sync"
)

const bucketSize = 20


// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	me      Contact
	buckets [IDLength * 8]*bucket
	mutex sync.RWMutex
}

// NewRoutingTable returns a new instance of a RoutingTable
func NewRoutingTable(me Contact) *RoutingTable {
	routingTable := &RoutingTable{}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = newBucket()
	}
	routingTable.me = me
	return routingTable
}

// AddContact add a new contact to the correct Bucket
func (routingTable *RoutingTable) AddContact(contact Contact) {
	bucketIndex := routingTable.getBucketIndex(contact.ID)

	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()


	bucket := routingTable.buckets[bucketIndex]
	bucket.AddContact(contact)
}


// Sort by contacts distance to each other, in ascending order.
type ByDistance []Contact
func (a ByDistance) Len() int           { return len(a) }
func (a ByDistance) Less(i, j int) bool { return a[i].distance.Less(a[j].distance) }
func (a ByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }


// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	var candidates []Contact
	bucketIndex := routingTable.getBucketIndex(target)

	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()

	bucket := routingTable.buckets[bucketIndex]

	candidates = append(candidates, bucket.GetContactAndCalcDistance(target)...)

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && len(candidates) < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates = append(candidates, bucket.GetContactAndCalcDistance(target)...)
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates = append(candidates, bucket.GetContactAndCalcDistance(target)...)
		}
	}

	sort.Sort(ByDistance(candidates))

	if count > len(candidates) {
		count = len(candidates)
	}

	return candidates[:count]
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()

	distance := id.CalcDistance(routingTable.me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}


func (routingTable *RoutingTable) GetMe() Contact {
	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()

	return routingTable.me
}


func (routingTable *RoutingTable) IsBucketFull(bucketIndex int) bool {
	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()

	return routingTable.buckets[bucketIndex].Full()
}

func (routingTable *RoutingTable) RemoveContactFromBucket(bucketIndex int, contact Contact) {
	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()

	routingTable.buckets[bucketIndex].RemoveContact(contact)
}

func (routingTable *RoutingTable) GetLastInBucket(bucketIndex int) Contact {
	routingTable.mutex.Lock()
	defer routingTable.mutex.Unlock()

	return routingTable.buckets[bucketIndex].GetLast()
}

