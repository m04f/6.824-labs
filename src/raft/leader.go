package raft

import (
	"context"
	"time"

	"6.5840/labrpc"
)

func (rf *Raft) sendHeartBeats(ctx context.Context, peer *labrpc.ClientEnd, term int) {
	sendHB := func() {
		var reply AppendEntriesReply
		if ok := peer.Call("Raft.AppendEntries", &AppendEntriesArgs{term}, &reply); ok {
			if reply.Term > term {
				rf.setTerm <- reply.Term
			}
		}
	}
	go sendHB()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Millisecond * heartBeatDelay):
			go sendHB()
		}
	}
}

func (rf *Raft) leader(ctx context.Context, term int) state {
	rf.isLeader.Store(true)
	defer rf.isLeader.Store(false)

	// start heartbeat go-routines
	childCtx, _ := rf.onTermChange(ctx, term)
	for i, peer := range rf.peers {
		if i == rf.me {
			continue
		}
		go rf.sendHeartBeats(childCtx, peer, term)
	}
	<-childCtx.Done()
	return rf.follower
}
