namespace go interview

struct CreateInterviewReq {
    1: required string candidate_id
    2: required string position
    3: required string direction
}

struct CreateInterviewResp {
    1: required string interview_id
    2: required string stage
    3: required string created_at
}

struct SubmitTurnReq {
    1: required string interview_id
    2: required string candidate_id
    3: required string text
}

struct SubmitTurnResp {
    1: required string reply
    2: required string stage
    3: required bool   is_finished
}

struct FinishInterviewReq {
    1: required string interview_id
}

struct FinishInterviewResp {
    1: required string finished_at
    2: required i64    duration_seconds
}

service InterviewService {
    CreateInterviewResp CreateInterview(1: CreateInterviewReq req)
    SubmitTurnResp      SubmitTurn(1: SubmitTurnReq req)
    FinishInterviewResp FinishInterview(1: FinishInterviewReq req)
}
