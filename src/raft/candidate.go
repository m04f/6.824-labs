package raft

import (
	"context"
	"time"

	"6.5840/labrpc"
)

func (rf *Raft) sendRequestVote(ctx context.Context, peer *labrpc.ClientEnd, term int, voteGranted chan<- bool) {
	args := RequestVoteArgs{term, rf.me}
	reply := RequestVoteReply{}
	ok := peer.Call("Raft.RequestVote", &args, &reply)
	if ok {
		voteGranted <- reply.VoteGranted
		rf.setTerm <- reply.Term
		return
	}
	select {
	case <-ctx.Done():
		voteGranted <- false
		return
	case <-time.After(time.Millisecond * 5):
		rf.sendRequestVote(ctx, peer, term, voteGranted)
	}
}

func (rf *Raft) voteCollector(term int, votes chan bool, promote chan<- struct{}) {
	votesGranted := 0
	votesReceived := 0
	requestedFromSelf := false
	promoted := false
	for voteGranted := range votes {
		votesReceived++
		if votesReceived == len(rf.peers) {
			close(votes)
		}
		if voteGranted {
			votesGranted++
		}
		if votesGranted == len(rf.peers)/2 && !requestedFromSelf {
			requestedFromSelf = true
			rf.voteRequests <- VoteRequest{term, rf.me, votes}
		}
		if votesGranted > len(rf.peers)/2 && !promoted {
			promoted = true
			close(promote)
		}
	}
}

func (rf *Raft) candidate(ctx context.Context, term int) state {

	rf.setTerm <- term
	promote := make(chan struct{})
	votes := make(chan bool)

	go rf.voteCollector(term, votes, promote)

	childCtx, cancel := rf.onTermChange(ctx, term)

	// track new leaders
	go func() {
		for {
			select {
			case <-childCtx.Done():
				return
			case hbTerm := <-rf.heartBeats:
				if hbTerm >= term {
					cancel()
					return
				}
			}
		}
	}()

	for i, peer := range rf.peers {
		if i == rf.me {
			continue
		}
		go rf.sendRequestVote(childCtx, peer, term, votes)
	}

	select {
	case <-promote:
		return func(ctx context.Context) state { return rf.leader(ctx, term) }
	case <-childCtx.Done():
		return rf.follower
	case <-rf.electionTimeout():
		return func(ctx context.Context) state { return rf.candidate(ctx, term+1) }
	}
}
