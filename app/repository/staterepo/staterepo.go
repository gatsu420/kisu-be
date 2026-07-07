package staterepo

func (r *repositoryImpl) Save(state string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.states[state] = struct{}{}
}

func (r *repositoryImpl) CheckExistence(state string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.states[state]
	if !ok {
		return false
	}

	return true
}
