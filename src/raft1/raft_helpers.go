// Helper functions for the Raft implementation.
// These are not thread-safe and must be called with appropriate locks held.
package raft

import "6.5840/labrpc"

// checkTerm compares the node's current term with a given term. If the given
// term is greater, the node's term is updated, its vote is reset, and it
// returns true. It returns the latest term and whether an update occurred.
func (rf *Raft) checkTerm(term int) (latestTerm int, updated bool) {
	if term > rf.term {
		rf.term = term
		rf.votedFor = -1
		return term, true
	}

	return rf.term, false
}

func (rf *Raft) isLeader() bool {
	return rf.votedFor == rf.me
}

func (rf *Raft) callRequestVote(server *labrpc.ClientEnd, reply *RequestVoteReply) bool {
	args := &RequestVoteArgs{
		Term:        rf.term,
		CandidateId: rf.me,
	}
	ok := server.Call("Raft.RequestVote", args, reply)
	return ok
}

func (rf *Raft) callAppendEntries(server *labrpc.ClientEnd, reply *AppendEntriesReply) bool {
	args := &AppendEntriesArgs{
		Term:     rf.term,
		LeaderId: rf.me,
	}
	ok := server.Call("Raft.AppendEntries", args, reply)
	return ok
}
