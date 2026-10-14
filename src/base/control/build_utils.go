package control

// GetProjectCommitHash returns the main project commit hash from BuildInfo.
// Returns the full hash string, or "unknown" if not found or struct not yet populated.
func GetProjectCommitHash() string {
	if len(BuildInfo.Repository.Commit) >= 40 {
		return BuildInfo.Repository.Commit
	}
	return "unknown"
}

// GetProjectBranch returns the main project branch name from BuildInfo.
func GetProjectBranch() string {
	if BuildInfo.Repository.Branch != "" {
		return BuildInfo.Repository.Branch
	}
	return "unknown"
}

// GetBuildDate returns the build timestamp from BuildInfo.
func GetBuildDate() string {
	if BuildInfo.BuildDate != "" {
		return BuildInfo.BuildDate
	}
	return "unknown"
}

// GetBuildUser returns the build user from BuildInfo.
func GetBuildUser() string {
	if BuildInfo.BuildUser != "" {
		return BuildInfo.BuildUser
	}
	return "unknown"
}

// GetBuildHostname returns the build hostname from BuildInfo.
func GetBuildHostname() string {
	if BuildInfo.BuildHostname != "" {
		return BuildInfo.BuildHostname
	}
	return "unknown"
}
