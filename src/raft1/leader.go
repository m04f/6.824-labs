package raft

import "time"

const heartbeatDelay = 150

func (rf *Raft) sendHbs(terms chan int) {

	if !rf.isLeader() {
		return
	}

	term := rf.term
	for i, peer := range rf.peers {
		if i == rf.me {
			continue
		}
		go func() {
			reply := &AppendEntriesReply{}
			rf.callAppendEntries(peer, reply)
			if reply.Term > term {
				terms <- reply.Term
			}
		}()
	}
}
func (rf *Raft) LeaderCoroutine() {
	terms := make(chan int)
	timer := time.After(time.Millisecond * heartbeatDelay)
	for !rf.killed() {
		select {
		case term := <-terms:
			func() {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				rf.checkTerm(term)
			}()
		case <-rf.becomeLeader:
			timer = time.After(time.Millisecond * heartbeatDelay)
			func() {
				rf.mu.RLock()
				defer rf.mu.RUnlock()
				rf.sendHbs(terms)
			}()
		case <-timer:
			timer = time.After(time.Millisecond * heartbeatDelay)
			func() {
				rf.mu.RLock()
				defer rf.mu.RUnlock()
				rf.sendHbs(terms)
			}()
		}
	}
}
