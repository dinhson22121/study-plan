# Phase 5 — Activity Loop Plan (quiz → progress → analytics)

> Detailed plan for the Phase 5 modules. Build order follows the data flow:
> **quiz** produces results → **progress** reacts (mastery, streaks, achievements)
> → **analytics** aggregates for dashboards and re-engagement.

## Data flow overview

```
student takes quiz ──► quiz.completed (eventbus)
                          │
                          ▼
                    progress: update TopicProgress + Streak
                          │  (new achievement?)
                          ├─► deps.Notifier ─► ACHIEVEMENT push (notification pipeline)
                          └─► achievement.unlocked (eventbus)
                          │
analytics (reads progress + quiz + curriculum) ──► dashboard, weak topics, inactivity
                          │
                          └─► feeds notification re-engagement scheduler (inactive users)
```

Reuses existing infrastructure: question bank (assemble/grade), `eventbus` (in-process
events), `deps.Notifier` (achievement + re-engagement pushes), curriculum (topic titles).

---

## Module 1 — quiz

Practice quizzes per topic: assemble from the question bank, grade, reveal a
per-question review (correct answer + explanation), and emit a completion event.

- **Domain**
  - `QuizSession` (aggregate): id, userID, topicID, questionIDs, status (IN_PROGRESS/COMPLETED), createdAt
  - `QuizResult`: sessionID, userID, topicID, score %, correctCount, total, passed, completedAt
  - `QuestionReview`: per-question outcome (selected, correct option, isCorrect, explanation) — shown only after submit
  - Pass threshold constant (decision below) → `passed`
- **Ports**
  - `Repository`: SaveSession, GetSession, MarkCompleted, SaveResult, GetResult, ListResultsByUser
  - `QuestionSource`: `SampleForTopic(topicID, limit)`, `Detail(questionIDs)` (stem, options, correct set, explanation) — adapter over `question.Service`
  - `EventPublisher` = `deps.Bus`
- **Flow**
  - `StartQuiz(userID, topicID, n)` → sample question ids → save session
  - `SubmitQuiz(sessionID, userID, answers)` → guard ownership/status → grade → save result + reviews → publish `quiz.completed` → return result **with review revealed**
- **Endpoints (JWT):** `POST /quizzes`, `POST /quizzes/:id/submit`, `GET /quizzes/:id`, `GET /quizzes`
- **Migration `010`:** `quiz_session`, `quiz_session_question`, `quiz_result`, `quiz_answer`
- **Event:** `quiz.completed` → `QuizCompletedEvent{userID, topicID, score, passed}`
- **Tests:** grading, pass threshold, review revealed only post-submit, ownership/double-submit guards, repo integration

## Module 2 — progress

Turns quiz activity into per-topic mastery, streaks, and achievements; fires
achievement notifications.

- **Domain**
  - `TopicProgress`: userID, topicID, status (NOT_STARTED/IN_PROGRESS/COMPLETED), bestScore, attempts, updatedAt — COMPLETED when bestScore ≥ mastery threshold
  - `Streak`: userID, currentStreak, longestStreak, lastActiveDate (increment on consecutive day, reset on gap)
  - `Achievement`: userID, type, ref, unlockedAt (recorded once → de-dupes notifications)
- **Ports**
  - `Repository`: UpsertTopicProgress, GetTopicProgress, ListProgressByUser, GetStreak, UpsertStreak, HasAchievement, SaveAchievement
  - `TopicTitleSource`: `Title(topicID)` — adapter over curriculum (for the ACHIEVEMENT `{topic}` variable)
  - `deps.Notifier` for achievement pushes
- **Event handling** — subscribe to `quiz.completed`:
  - update TopicProgress (attempts++, bestScore = max, recompute status)
  - update Streak (today vs lastActiveDate)
  - detect **new** achievements → SaveAchievement + `Notifier.EnqueueReminder(ACHIEVEMENT, ACHIEVEMENT_V1, {topic})` with per-achievement idempotency key; emit `achievement.unlocked`
- **Endpoints (JWT):** `GET /progress` (streak + counts), `GET /progress/topics`
- **Migration `011`:** `topic_progress`, `streak`, `achievement`
- **Tests:** mastery transitions, streak increment/reset, achievement de-dup, notification enqueued, event handler

## Module 3 — analytics

Read-model dashboards + the inactivity feed for re-engagement.

- **Domain / read models:** `DashboardSnapshot` (topic completion %, subject completion, quiz average, streak, last-active), `WeakTopic` list
- **Ports**
  - reads from progress + quiz + curriculum services (composition adapters)
  - optional `activity_event` log for last-active / inactivity (decision below)
- **Use cases**
  - `Dashboard(userID)` → aggregates the metrics above
  - `InactiveUsers(days)` + `WeakestTopic(userID)` → consumed by the notification re-engagement scheduler (currently a stub)
- **Endpoints (JWT):** `GET /analytics/me`, `GET /analytics/me/weak-topics`
- **Migration `012`:** `activity_event` (only if event-fed approach chosen)
- **Tests:** aggregation math, weak-topic ranking, inactivity window

---

## Decisions (locked)

1. **Mastery threshold = 80%.** A topic is COMPLETED when best quiz score ≥ 80%. Quizzes are topic-scoped; the subject-scoped "weekly quiz" is **deferred** (the notification scheduler already nudges weekly quizzes).
2. **Achievements:** `TOPIC_COMPLETED`, `STREAK_7`, `STREAK_30`, `PERFECT_SCORE` (100%). Each recorded once → de-duped pushes.
3. **Analytics = hybrid.** Completion %, subject completion, and quiz average are computed on-demand from progress + quiz; a small `activity_event` log (appended on `quiz.completed`) powers last-active / inactivity.
4. **Drop "Study Hours"** (no time-tracking source yet) and **complete re-engagement** in Phase 5.

### Re-engagement wiring (closes PRD §4 re-engagement)

- New port `app.ReengagementSource { InactiveUserIDs(ctx, days) ([]string, error) }`; **analytics** sets `deps.ReengagementSource` on Register (computed from `activity_event`).
- The **notification** re-engagement cron (currently a stub) reads `deps.ReengagementSource`, lists users inactive ≥ 3 days, and enqueues `REENGAGEMENT` (template `REENGAGEMENT_V1`, `{days}`) per user through the dispatcher (preference gate + idempotency per user/day).
- Weak-topic personalization of the re-engagement copy stays a future enhancement.

### Module registration order (main.go)

`auth → user → curriculum → question → content → goal → placement → quiz → analytics → notification → progress → studyplan`

- `analytics` before `notification` (notification's re-engagement cron reads `deps.ReengagementSource`, set by analytics).
- `progress` and `studyplan` after `notification` (they need `deps.Notifier`).
- `quiz` anywhere before runtime; `progress` subscribes to `quiz.completed` at registration, so its position relative to `quiz` doesn't matter (events fire only after startup).
- Only the `deps.*` port handoffs (Notifier, ReengagementSource, AuthValidate) constrain order.

### Migrations

`010` quiz · `011` progress · `012` analytics (`activity_event`).

## Cross-cutting (unchanged from prior phases)

DDD layering · cross-context via ports+adapters (no struct imports) · events over eventbus ·
`deps.Notifier` for pushes · TDD on domain/application, integration-tagged repo tests ·
register progress/analytics **after** notification (need `deps.Notifier`).
