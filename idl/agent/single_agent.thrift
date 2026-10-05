namespace go agent


// The type of transmission allowed by the context
enum ContextMode {
    Chat           = 0,
    FunctionCall_1 = 1,
    FunctionCall_2 = 2,
    FunctionCall_3 = 3,
}

enum ModelResponseFormat {
    Text = 0,
    Markdown = 1,
    JSON = 2,
}

enum ModelStyle {
    Custom = 0
    Creative = 1
    Balance = 2
    Precise = 3
}

enum CacheType {
    CacheClosed = 0 // 缓存关闭
    PrefixCache = 1 // 前缀缓存
}


struct ShortMemoryPolicy {
    1: optional ContextMode ContextMode  (api.body="context_mode") , // The type of transmission allowed by the context
    2: optional i32         HistoryRound (api.body="history_round"), // Number of rounds of context band
}



struct ModelInfo {
    1: optional i64                 ModelId           (agw.js_conv="str", api.js_conv="true", api.body="model_id"), // Model ID
    2: optional double              Temperature       (api.body="temperature")                                    , // Temperature, model output randomness, the larger the value, the more random, the smaller the more conservative (0-1]
    3: optional i32                 MaxTokens         (api.body="max_tokens")                                     , // Maximum Token Reply
    4: optional double              TopP              (api.body="top_p")                                          , // Another model's output randomness, the larger the value, the more random [0, 1]
    5: optional double              FrequencyPenalty  (api.body="frequency_penalty")                              , // Frequency penalty, adjust the frequency of words in the generated content, the fewer positive words are [-1.0, 1.0]
    6: optional double              PresencePenalty   (api.body="presence_penalty")                               , // There is a penalty, adjust the frequency of new words in the generated content, avoid repeating words with positive values, and use new words [-1.0, 1.0]
    7: optional ShortMemoryPolicy   ShortMemoryPolicy (api.body="short_memory_policy")                            , // contextual policy
    8: optional i32                 TopK              (api.body="top_k")                                          , // When generating, sample the size of the candidate set
    9: optional ModelResponseFormat ResponseFormat    (api.body="response_format")                                , // model reply content format
    10: optional ModelStyle         ModelStyle        (api.body="model_style")
    11: optional CacheType          CacheType           (api.body="cache_type")                                   , // 缓存配置
    12: optional bool               SpCurrentTime       (api.body="sp_current_time")                                , // sp拼接当前时间
    13: optional bool               SpAntiLeak          (api.body="sp_anti_leak")                                   , // sp拼接防泄露指令
    14: optional bool               SpVoiceInfo         (api.body="sp_voice_info")                                  , // sp拼接声纹信息
    15: optional map<string, string> Parameters         (api.body="parameters")                                   , // User-selected model style
}

struct PromptInfo {
    1: optional string Prompt (api.body="prompt"), // Text prompt
}

struct Knowledge{
    1: list<string> Currency

}

struct AgentDraft{
    1: i64 Id
    2: string Name
    3: string Description
    4: string IconUri
    5: optional ModelInfo ModelInfo (api.body="model_info")                         // model configuration
    6: optional PromptInfo PromptInfo (api.body="prompt_info")                      // Prompt message
    7: optional Knowledge Knowledge
}

struct AgentDraftResponse{
    1: list<AgentDraft> items
}



service SingleAgentController{
    AgentDraftResponse GetPageList() (api.get= "/agent/list")


}