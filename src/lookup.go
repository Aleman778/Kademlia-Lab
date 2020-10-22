package main

import (
	"sort"
	"sync"
)

type Lookup struct {
	id *KademliaID

	notCalled []Contact
	called []Contact

	wg sync.WaitGroup

	sendCh chan<- Contact
	receiveCh <-chan LookupResponse

	closestHasNotValue Contact

	mutex sync.RWMutex
}

type LookupResponse struct {
	contacts []Contact
	contact Contact
	hasValue bool
}

func RunLookup(id *KademliaID, me Contact, contacts []Contact, sendCh chan<- Contact, receiveCh <-chan LookupResponse) ([]Contact, Contact, int) {
	me.CalcDistance(id)
	lookup := Lookup{
		id: id,
		notCalled: contacts,
		called: append([]Contact{}, me),
		wg: sync.WaitGroup{},
		sendCh: sendCh,
		receiveCh: receiveCh,
		closestHasNotValue: me,
		mutex: sync.RWMutex{}}

	lookup.wg.Add(1)
	lookup.recursiveLookup()

	lookup.wg.Wait()

    before := 0
    for i, contact := range lookup.called {
        if contact.ID.Equals(lookup.closestHasNotValue.ID) {
            before = i
            break;
        }
    }

	if len(lookup.called) < k {
		return lookup.called, lookup.closestHasNotValue, before
	}
	return lookup.called[:k], lookup.closestHasNotValue, before
}

func (self *Lookup) recursiveLookup() {
	defer self.wg.Done()
	noNewContacts := true
	isDone := false

	numRunning := 0
	for i := 0; i < ALPHA; i++ {
		if len(self.notCalled) != 0 {
			self.callContact()
			numRunning += 1
		} else {
			break
		}
	}

	for numRunning != 0 {
		newContacts := self.getResponse()
		numRunning -= 1

		if len(self.notCalled) != 0 && len(self.called) >= k  {
			isDone = self.called[k-1].distance.Less(self.notCalled[0].distance)
		}

		if len(newContacts) > 0 && !isDone {
			noNewContacts = false
			self.wg.Add(1)
			self.recursiveLookup()
		}
	}

	if noNewContacts && !isDone {
		self.lastEffort()
	}
}

func (self *Lookup) lastEffort() {
	numRunning := 0
	for i := 0; i < k; i++ {
		if len(self.notCalled) != 0 {
			self.callContact()
			numRunning += 1
		} else {
			break
		}
	}

	for numRunning != 0 {
		self.getResponse()
		numRunning -= 1
	}
	return
}

func (self *Lookup) callContact() {
	self.mutex.Lock()

	var contact Contact
	self.notCalled, contact = PopCandidate(self.notCalled)
	contact.CalcDistance(self.id)
	self.called = append(self.called, contact)
	sort.Sort(ByDistance(self.called))
	self.sendCh <- contact

	self.mutex.Unlock()
}

func (self *Lookup) getResponse() []Contact {
	response := <-self.receiveCh

	if !response.hasValue && response.contact.ID.Less(self.closestHasNotValue.ID) {
		self.closestHasNotValue = response.contact
	}

	var newContacts []Contact
	self.mutex.Lock()
	// Add uncontacted nodes to the notCalled list.
	for _, contact := range response.contacts {
		if !InCandidates(self.notCalled, contact) && !InCandidates(self.called, contact) {
			contact.CalcDistance(self.id)
			self.notCalled = append(self.notCalled, contact)
			newContacts = append(newContacts, contact)
		}
	}
	sort.Sort(ByDistance(self.notCalled))
	self.mutex.Unlock()

	return newContacts
}

