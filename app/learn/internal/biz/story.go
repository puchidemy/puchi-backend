package biz

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/puchidemy/puchi-backend/app/learn/internal/data/sqlc/gen"
)

const defaultStoryXP int32 = 20

var (
	ErrActivityNotFound  = errors.New("activity not found")
	ErrActivityForbidden = errors.New("activity forbidden")
	ErrSceneIncomplete   = errors.New("scene activities incomplete")
)

// StoryRepoInterface reads story-first curriculum rows.
type StoryRepoInterface interface {
	ListCities(ctx context.Context) ([]gen.LearnCity, error)
	GetCityBySlug(ctx context.Context, slug string) (*gen.LearnCity, error)
	GetCityByID(ctx context.Context, id string) (*gen.LearnCity, error)
	CountPublishedStoriesByCity(ctx context.Context, cityID string) (int32, error)
	CountCompletedStoriesByOwnerCity(ctx context.Context, ownerType, ownerID, cityID string) (int32, error)
	ListPublishedStoriesByCity(ctx context.Context, cityID string) ([]gen.LearnStory, error)
	GetStoryByID(ctx context.Context, id string) (*gen.LearnStory, error)
	ListScenesByStoryID(ctx context.Context, storyID string) ([]gen.LearnScene, error)
	GetSceneByID(ctx context.Context, id string) (*gen.LearnScene, error)
	ListActivitiesBySceneID(ctx context.Context, sceneID string) ([]gen.LearnActivity, error)
	GetActivityByID(ctx context.Context, id string) (*gen.LearnActivity, error)
	ListActivitiesByStoryID(ctx context.Context, storyID string) ([]gen.LearnActivity, error)
}

// CityListItem is a city with story counts for the journey map.
type CityListItem struct {
	City                gen.LearnCity
	StoryCount          int32
	CompletedStoryCount int32
}

// StorySummaryItem is a published story with owner progress.
type StorySummaryItem struct {
	Story          gen.LearnStory
	ProgressStatus string
}

// CityDetail is a city hub payload.
type CityDetail struct {
	City                 CityListItem
	Stories              []StorySummaryItem
	ContinueStoryID      string
	RecommendedStoryIDs  []string
}

// ActivityPrompt is an activity without answer keys.
type ActivityPrompt struct {
	Activity gen.LearnActivity
}

// SceneWithActivities is a scene with activities and progress.
type SceneWithActivities struct {
	Scene          gen.LearnScene
	ProgressStatus string
	Activities     []ActivityPrompt
}

// StoryDetail is a full story player payload.
type StoryDetail struct {
	Story          gen.LearnStory
	CitySlug       string
	ProgressStatus string
	Scenes         []SceneWithActivities
}

// CompleteSceneResult is returned after marking a scene complete.
type CompleteSceneResult struct {
	SceneCompleted      bool
	StoryCompleted      bool
	CompletedSceneCount int32
	SoftGate            bool
}

// StoryCompletionSummary is shown on the story completion screen.
type StoryCompletionSummary struct {
	VocabFocus       []string
	GrammarFocus     []string
	ListeningSeconds int32
	CulturalFact     string
}

// CompleteStoryResult is returned after marking a story complete.
type CompleteStoryResult struct {
	XP             int32
	StoryCompleted bool
	Summary        StoryCompletionSummary
}

func (uc *LearnUsecase) countCompletedScenes(ctx context.Context, ownerType, ownerID string) (int32, error) {
	return uc.progressRepo.CountCompletedScenesByOwner(ctx, ownerType, ownerID)
}

// assertGuestSceneSoftGate blocks guests from starting/completing new scenes after 3 completions.
func (uc *LearnUsecase) assertGuestSceneSoftGate(ctx context.Context, ownerType, ownerID, sceneID string) error {
	if ownerType != "guest" {
		return nil
	}

	progress, err := uc.progressRepo.GetSceneProgress(ctx, ownerType, ownerID, sceneID)
	if err == nil && progress.Status == "completed" {
		return nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	n, err := uc.countCompletedScenes(ctx, ownerType, ownerID)
	if err != nil {
		return err
	}
	if n < int32(guestSoftGateCompletedLimit) {
		return nil
	}
	return ErrGuestSoftGate
}

// ListCities returns all cities with published/completed story counts.
func (uc *LearnUsecase) ListCities(ctx context.Context, ownerType, ownerID string) ([]CityListItem, error) {
	cities, err := uc.storyRepo.ListCities(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]CityListItem, 0, len(cities))
	for _, city := range cities {
		storyCount, err := uc.storyRepo.CountPublishedStoriesByCity(ctx, city.ID)
		if err != nil {
			return nil, err
		}
		completed := int32(0)
		if ownerType != "" && ownerID != "" {
			completed, err = uc.storyRepo.CountCompletedStoriesByOwnerCity(ctx, ownerType, ownerID, city.ID)
			if err != nil {
				return nil, err
			}
		}
		out = append(out, CityListItem{
			City:                city,
			StoryCount:          storyCount,
			CompletedStoryCount: completed,
		})
	}
	return out, nil
}

// GetCity returns a city hub with published stories and recommendations.
func (uc *LearnUsecase) GetCity(ctx context.Context, ownerType, ownerID, slug string) (*CityDetail, error) {
	city, err := uc.storyRepo.GetCityBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}

	storyCount, err := uc.storyRepo.CountPublishedStoriesByCity(ctx, city.ID)
	if err != nil {
		return nil, err
	}
	completedCount, err := uc.storyRepo.CountCompletedStoriesByOwnerCity(ctx, ownerType, ownerID, city.ID)
	if err != nil {
		return nil, err
	}

	stories, err := uc.storyRepo.ListPublishedStoriesByCity(ctx, city.ID)
	if err != nil {
		return nil, err
	}

	progressByID := make(map[string]string)
	progressRows, err := uc.progressRepo.ListStoryProgressByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return nil, err
	}
	for _, row := range progressRows {
		progressByID[row.StoryID] = row.Status
	}

	summaries := make([]StorySummaryItem, 0, len(stories))
	var continueID string
	var recommended []string
	for _, story := range stories {
		status := progressByID[story.ID]
		if status == "" {
			status = "not_started"
		}
		summaries = append(summaries, StorySummaryItem{Story: story, ProgressStatus: status})
		if continueID == "" && status == "in_progress" {
			continueID = story.ID
		}
		if status != "completed" {
			recommended = append(recommended, story.ID)
		}
	}
	if len(recommended) > 3 {
		recommended = recommended[:3]
	}

	return &CityDetail{
		City: CityListItem{
			City:                *city,
			StoryCount:          storyCount,
			CompletedStoryCount: completedCount,
		},
		Stories:             summaries,
		ContinueStoryID:     continueID,
		RecommendedStoryIDs: recommended,
	}, nil
}

// GetStory returns a story with scenes and activity prompts (no answer keys).
func (uc *LearnUsecase) GetStory(ctx context.Context, ownerType, ownerID, storyID string) (*StoryDetail, error) {
	story, err := uc.storyRepo.GetStoryByID(ctx, storyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}
	if story.Status != "published" {
		return nil, ErrCurriculumNotFound
	}

	city, err := uc.storyRepo.GetCityByID(ctx, story.CityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}

	storyStatus := "not_started"
	if sp, err := uc.progressRepo.GetStoryProgress(ctx, ownerType, ownerID, storyID); err == nil {
		storyStatus = sp.Status
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	scenes, err := uc.storyRepo.ListScenesByStoryID(ctx, storyID)
	if err != nil {
		return nil, err
	}

	sceneProgressByID := make(map[string]string)
	sceneRows, err := uc.progressRepo.ListSceneProgressByOwner(ctx, ownerType, ownerID)
	if err != nil {
		return nil, err
	}
	for _, row := range sceneRows {
		sceneProgressByID[row.SceneID] = row.Status
	}

	outScenes := make([]SceneWithActivities, 0, len(scenes))
	for _, scene := range scenes {
		activities, err := uc.storyRepo.ListActivitiesBySceneID(ctx, scene.ID)
		if err != nil {
			return nil, err
		}
		prompts := make([]ActivityPrompt, 0, len(activities))
		for _, a := range activities {
			prompts = append(prompts, ActivityPrompt{Activity: a})
		}
		status := sceneProgressByID[scene.ID]
		if status == "" {
			status = "not_started"
		}
		outScenes = append(outScenes, SceneWithActivities{
			Scene:          scene,
			ProgressStatus: status,
			Activities:     prompts,
		})
	}

	return &StoryDetail{
		Story:          *story,
		CitySlug:       city.Slug,
		ProgressStatus: storyStatus,
		Scenes:         outScenes,
	}, nil
}

// StartActivity creates or resumes an activity attempt for a scene.
func (uc *LearnUsecase) StartActivity(ctx context.Context, ownerType, ownerID, sceneID string) (uuid.UUID, error) {
	scene, err := uc.storyRepo.GetSceneByID(ctx, sceneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrCurriculumNotFound
		}
		return uuid.Nil, err
	}
	story, err := uc.storyRepo.GetStoryByID(ctx, scene.StoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrCurriculumNotFound
		}
		return uuid.Nil, err
	}
	if story.Status != "published" {
		return uuid.Nil, ErrCurriculumNotFound
	}

	if err := uc.assertGuestSceneSoftGate(ctx, ownerType, ownerID, sceneID); err != nil {
		return uuid.Nil, err
	}

	if existing, err := uc.attemptRepo.GetActiveActivityAttemptByOwnerScene(ctx, ownerType, ownerID, sceneID); err == nil {
		id, parseErr := uuid.Parse(existing.ID)
		if parseErr != nil {
			return uuid.Nil, parseErr
		}
		return id, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	attempt, err := uc.attemptRepo.CreateActivityAttempt(ctx, ownerType, ownerID, story.ID, sceneID)
	if err != nil {
		return uuid.Nil, err
	}
	if err := uc.progressRepo.UpsertStoryProgress(ctx, ownerType, ownerID, story.ID, "in_progress", 0); err != nil {
		return uuid.Nil, err
	}
	if err := uc.progressRepo.UpsertSceneProgress(ctx, ownerType, ownerID, sceneID, "in_progress"); err != nil {
		return uuid.Nil, err
	}

	id, err := uuid.Parse(attempt.ID)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// SubmitActivityAnswer grades a payload against an activity answer key.
func (uc *LearnUsecase) SubmitActivityAnswer(ctx context.Context, ownerType, ownerID, attemptID, activityID string, payload json.RawMessage) (bool, error) {
	attempt, err := uc.attemptRepo.GetActivityAttemptByID(ctx, attemptID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrAttemptNotFound
		}
		return false, err
	}
	if attempt.OwnerType != ownerType || attempt.OwnerID != ownerID {
		return false, ErrAttemptForbidden
	}
	if attempt.Status != "active" {
		return false, ErrAttemptNotActive
	}

	activity, err := uc.storyRepo.GetActivityByID(ctx, activityID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrActivityNotFound
		}
		return false, err
	}
	if activity.SceneID != attempt.SceneID {
		return false, ErrActivityForbidden
	}

	correct, err := Grade(activity.Type, activity.Prompt, activity.Answer, payload)
	if err != nil {
		return false, err
	}
	if err := uc.attemptRepo.InsertActivityAttemptAnswer(ctx, attemptID, activityID, payload, correct); err != nil {
		return false, err
	}
	return correct, nil
}

// CompleteScene finalizes the active attempt and marks the scene completed.
func (uc *LearnUsecase) CompleteScene(ctx context.Context, ownerType, ownerID, sceneID string) (*CompleteSceneResult, error) {
	scene, err := uc.storyRepo.GetSceneByID(ctx, sceneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}
	if err := uc.assertGuestSceneSoftGate(ctx, ownerType, ownerID, sceneID); err != nil {
		return nil, err
	}

	activities, err := uc.storyRepo.ListActivitiesBySceneID(ctx, sceneID)
	if err != nil {
		return nil, err
	}

	// Idempotent: already completed scene.
	if prev, err := uc.progressRepo.GetSceneProgress(ctx, ownerType, ownerID, sceneID); err == nil && prev.Status == "completed" {
		count, err := uc.countCompletedScenes(ctx, ownerType, ownerID)
		if err != nil {
			return nil, err
		}
		storyCompleted, err := uc.isStoryFullyCompleted(ctx, ownerType, ownerID, scene.StoryID)
		if err != nil {
			return nil, err
		}
		return &CompleteSceneResult{
			SceneCompleted:      true,
			StoryCompleted:      storyCompleted,
			CompletedSceneCount: count,
			SoftGate:            ownerType == "guest" && count >= int32(guestSoftGateCompletedLimit),
		}, nil
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	attempt, err := uc.attemptRepo.GetActiveActivityAttemptByOwnerScene(ctx, ownerType, ownerID, sceneID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAttemptNotFound
		}
		return nil, err
	}

	answers, err := uc.attemptRepo.ListActivityAttemptAnswersByAttemptID(ctx, attempt.ID)
	if err != nil {
		return nil, err
	}
	if !allActivitiesAnswered(activities, answers) {
		return nil, ErrSceneIncomplete
	}

	const sceneXPReward int32 = 5
	sessionXP := computeSessionXP(sceneXPReward, len(activities), countCorrectActivityAnswers(answers))
	if err := uc.attemptRepo.CompleteActivityAttempt(ctx, attempt.ID, sessionXP); err != nil {
		return nil, err
	}
	if err := uc.progressRepo.UpsertSceneProgress(ctx, ownerType, ownerID, sceneID, "completed"); err != nil {
		return nil, err
	}
	if err := uc.progressRepo.UpsertStoryProgress(ctx, ownerType, ownerID, scene.StoryID, "in_progress", 0); err != nil {
		return nil, err
	}

	storyCompleted, err := uc.markStoryProgressIfAllScenesComplete(ctx, ownerType, ownerID, scene.StoryID)
	if err != nil {
		return nil, err
	}

	count, err := uc.countCompletedScenes(ctx, ownerType, ownerID)
	if err != nil {
		return nil, err
	}

	if ownerType == "user" {
		if err := uc.publisher.PublishSceneCompleted(ctx, SceneCompletedEvent{
			UserID:      ownerID,
			SceneID:     sceneID,
			StoryID:     scene.StoryID,
			CompletedAt: time.Now().UTC(),
		}); err != nil {
			return nil, err
		}
	}

	return &CompleteSceneResult{
		SceneCompleted:      true,
		StoryCompleted:      storyCompleted,
		CompletedSceneCount: count,
		SoftGate:            ownerType == "guest" && count >= int32(guestSoftGateCompletedLimit),
	}, nil
}

// CompleteStory marks a story completed and returns a completion summary.
func (uc *LearnUsecase) CompleteStory(ctx context.Context, ownerType, ownerID, storyID string) (*CompleteStoryResult, error) {
	story, err := uc.storyRepo.GetStoryByID(ctx, storyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCurriculumNotFound
		}
		return nil, err
	}
	if story.Status != "published" {
		return nil, ErrCurriculumNotFound
	}

	scenes, err := uc.storyRepo.ListScenesByStoryID(ctx, storyID)
	if err != nil {
		return nil, err
	}
	for _, scene := range scenes {
		progress, err := uc.progressRepo.GetSceneProgress(ctx, ownerType, ownerID, scene.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrSceneIncomplete
			}
			return nil, err
		}
		if progress.Status != "completed" {
			return nil, ErrSceneIncomplete
		}
	}

	prevXP := int32(0)
	if prev, err := uc.progressRepo.GetStoryProgress(ctx, ownerType, ownerID, storyID); err == nil {
		prevXP = prev.XpEarned
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	xp := maxInt32(prevXP, defaultStoryXP)
	if err := uc.progressRepo.UpsertStoryProgress(ctx, ownerType, ownerID, storyID, "completed", xp); err != nil {
		return nil, err
	}

	fact := ""
	if story.CulturalFact != nil {
		fact = *story.CulturalFact
	}
	listening := int32(0)
	if story.EstMinutes != nil && *story.EstMinutes > 0 {
		listening = *story.EstMinutes * 60 / 2
	}

	if ownerType == "user" {
		completedAt := time.Now().UTC()
		if err := uc.publisher.PublishStoryCompleted(ctx, StoryCompletedEvent{
			UserID:      ownerID,
			StoryID:     storyID,
			CityID:      story.CityID,
			XP:          xp,
			CompletedAt: completedAt,
		}); err != nil {
			return nil, err
		}
		// Dual-publish legacy lesson event so Core XP/streak still apply during transition.
		if err := uc.publisher.PublishLessonCompleted(ctx, LessonCompletedEvent{
			UserID:      ownerID,
			LessonID:    storyID,
			UnitID:      story.CityID,
			XP:          xp,
			CompletedAt: completedAt,
		}); err != nil {
			return nil, err
		}
	}

	return &CompleteStoryResult{
		XP:             xp,
		StoryCompleted: true,
		Summary: StoryCompletionSummary{
			VocabFocus:       append([]string(nil), story.VocabFocus...),
			GrammarFocus:     append([]string(nil), story.GrammarFocus...),
			ListeningSeconds: listening,
			CulturalFact:     fact,
		},
	}, nil
}

func (uc *LearnUsecase) isStoryFullyCompleted(ctx context.Context, ownerType, ownerID, storyID string) (bool, error) {
	if sp, err := uc.progressRepo.GetStoryProgress(ctx, ownerType, ownerID, storyID); err == nil {
		return sp.Status == "completed", nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	return uc.markStoryProgressIfAllScenesComplete(ctx, ownerType, ownerID, storyID)
}

func (uc *LearnUsecase) markStoryProgressIfAllScenesComplete(ctx context.Context, ownerType, ownerID, storyID string) (bool, error) {
	scenes, err := uc.storyRepo.ListScenesByStoryID(ctx, storyID)
	if err != nil {
		return false, err
	}
	if len(scenes) == 0 {
		return false, nil
	}
	for _, scene := range scenes {
		progress, err := uc.progressRepo.GetSceneProgress(ctx, ownerType, ownerID, scene.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil
			}
			return false, err
		}
		if progress.Status != "completed" {
			return false, nil
		}
	}
	if err := uc.progressRepo.UpsertStoryProgress(ctx, ownerType, ownerID, storyID, "completed", defaultStoryXP); err != nil {
		return false, err
	}
	return true, nil
}

func allActivitiesAnswered(activities []gen.LearnActivity, answers []gen.LearnActivityAttemptAnswer) bool {
	if len(activities) == 0 {
		return true
	}
	seen := make(map[string]bool, len(answers))
	for _, a := range answers {
		seen[a.ActivityID] = true
	}
	for _, activity := range activities {
		if !seen[activity.ID] {
			return false
		}
	}
	return true
}

func countCorrectActivityAnswers(answers []gen.LearnActivityAttemptAnswer) int {
	// Prefer latest answer per activity for XP; count distinct correct activities.
	latest := make(map[string]bool)
	for _, a := range answers {
		latest[a.ActivityID] = a.Correct
	}
	n := 0
	for _, correct := range latest {
		if correct {
			n++
		}
	}
	return n
}
