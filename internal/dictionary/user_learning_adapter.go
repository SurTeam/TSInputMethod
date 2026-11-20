package dictionary

type LearningObserver struct {
	learner *PhraseLearner
	user    *UserDictionary
	save    bool
}

func NewLearningObserver(user *UserDictionary, minRepeats int, autoSave bool) *LearningObserver {
	return &LearningObserver{
		learner: NewPhraseLearner(user, minRepeats),
		user:    user,
		save:    autoSave,
	}
}

func (o *LearningObserver) ObserveCommittedPhrase(text string, syllables []string) error {
	if o == nil || o.learner == nil {
		return nil
	}
	_, added, err := o.learner.ObserveCommittedPhrase(text, syllables)
	if err != nil {
		return err
	}
	if added && o.save && o.user != nil {
		return o.user.Save()
	}
	return nil
}

func (o *LearningObserver) Learner() *PhraseLearner {
	if o == nil {
		return nil
	}
	return o.learner
}
