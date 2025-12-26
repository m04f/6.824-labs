package raft

import (
	"math/rand"
	"sync"
	"time"
)

func (rf *Raft) timeoutTimer() <-chan time.Time {
	ms := 150 + (rand.Int63() % 300)
	return time.After(time.Duration(ms) * time.Millisecond)
}

func (rf *Raft) CandidateCoroutine() {
	votes := make(chan int)
	terms := make(chan int)
	votesCount := 0
	votesTerm := -1
	votesLock := sync.Mutex{}
	votesTarget := len(rf.peers) / 2
	if len(rf.peers)%2 == 1 {
		votesTarget += 1
	}

	timer := rf.timeoutTimer()

	for !rf.killed() {
		select {
		case term := <-terms:
			func() {
				rf.mu.RLock()
				defer rf.mu.RUnlock()

				rf.checkTerm(term)
			}()
		case <-rf.heartBeats:
			timer = rf.timeoutTimer()
		case term := <-votes:
			func() {
				votesLock.Lock()
				defer votesLock.Unlock()

				if term < votesTerm {
					return
				}

				if term > votesTerm {
					votesTerm = term
					votesCount = 0
				}

				votesCount += 1

				if votesCount < votesTarget-1 {
					return
				}

				rf.mu.Lock()
				defer rf.mu.Unlock()

				if rf.term > votesTerm {
					votesTerm = rf.term
					votesCount = 0
					return
				}

				if rf.isLeader() {
					return
				}
				if rf.votedFor == -1 || votesCount == votesTarget {
					// become a leader
					// todo: verify that it's safe to use this
					rf.votedFor = rf.me
					rf.becomeLeader <- struct{}{}
					return
				}
			}()
		case <-timer:
			timer = rf.timeoutTimer()
			func() {
				rf.mu.Lock()
				defer rf.mu.Unlock()

				if rf.isLeader() {
					return
				}

				term, _ := rf.checkTerm(rf.term + 1)
				for i, peer := range rf.peers {
					if i == rf.me {
						continue
					}

					go func() {
						reply := &RequestVoteReply{}
						rf.callRequestVote(peer, term, reply)
						if reply.Term > term {
							terms <- reply.Term
							return
						}
						if reply.VoteGranted {
							votes <- term
							return
						}
					}()
				}
			}()
		}
	}
}
