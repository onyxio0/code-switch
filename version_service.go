package main

<<<<<<< HEAD
const AppVersion = "v0.1.8"
=======
const AppVersion = "v1.2.5"
>>>>>>> rogers/main

type VersionService struct {
	version string
}

func NewVersionService() *VersionService {
	return &VersionService{version: AppVersion}
}

func (vs *VersionService) CurrentVersion() string {
	return vs.version
}
