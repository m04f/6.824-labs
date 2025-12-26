package raft

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	term, updated := rf.checkTerm(args.Term)
	reply.Term = term
	reply.Success = updated || args.Term == reply.Term
	if reply.Success {
		rf.heartBeats <- struct{}{}
	}
}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
type RequestVoteArgs struct {
	// Term Candidate’s term
	Term int
	// CandidateId Candidate requesting vote
	CandidateId int
	// Your data here (3B).
}

type RequestVoteReply struct {
	// Term currentTerm, for candidate to update itself
	Term int
	// VoteGranted true means candidate received vote
	VoteGranted bool
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (3B).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	term, updated := rf.checkTerm(args.Term)
	reply.Term = term
	if updated || rf.votedFor == args.CandidateId {
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		return
	}

	reply.VoteGranted = false
}
