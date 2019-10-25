package webview

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/windmilleng/tilt/internal/cloud/cloudurl"
	"github.com/windmilleng/tilt/internal/dockercompose"
	"github.com/windmilleng/tilt/internal/ospath"
	"github.com/windmilleng/tilt/internal/store"
	"github.com/windmilleng/tilt/pkg/model"
	proto_webview "github.com/windmilleng/tilt/pkg/webview"
)

func StateToWebView(s store.EngineState) View {
	ret := View{}

	ret.Resources = append(ret.Resources, tiltfileResourceView(s))

	for _, name := range s.ManifestDefinitionOrder {
		mt, ok := s.ManifestTargets[name]
		if !ok {
			continue
		}

		ms := mt.State

		var absWatchDirs []string
		var absWatchPaths []string
		for _, p := range mt.Manifest.LocalPaths() {
			fi, err := os.Stat(p)
			if err == nil && !fi.IsDir() {
				absWatchPaths = append(absWatchPaths, p)
			} else {
				absWatchDirs = append(absWatchDirs, p)
			}
		}
		absWatchPaths = append(absWatchPaths, s.TiltfilePath)
		relWatchDirs := ospath.TryAsCwdChildren(absWatchDirs)
		relWatchPaths := ospath.TryAsCwdChildren(absWatchPaths)

		var pendingBuildEdits []string
		for _, status := range ms.BuildStatuses {
			for f := range status.PendingFileChanges {
				pendingBuildEdits = append(pendingBuildEdits, f)
			}
		}

		pendingBuildEdits = ospath.FileListDisplayNames(absWatchDirs, pendingBuildEdits)

		buildHistory := append([]model.BuildRecord{}, ms.BuildHistory...)
		for i, build := range buildHistory {
			build.Edits = ospath.FileListDisplayNames(absWatchDirs, build.Edits)
			buildHistory[i] = build
		}

		currentBuild := ms.CurrentBuild
		currentBuild.Edits = ospath.FileListDisplayNames(absWatchDirs, ms.CurrentBuild.Edits)

		// Sort the strings to make the outputs deterministic.
		sort.Strings(pendingBuildEdits)

		endpoints := store.ManifestTargetEndpoints(mt)

		podID := ms.MostRecentPod().PodID

		// NOTE(nick): Right now, the UX is designed to show the output exactly one
		// pod. A better UI might summarize the pods in other ways (e.g., show the
		// "most interesting" pod that's crash looping, or show logs from all pods
		// at once).
		hasPendingChanges, pendingBuildSince := ms.HasPendingChanges()
		r := Resource{
			Name:               name,
			DirectoriesWatched: relWatchDirs,
			PathsWatched:       relWatchPaths,
			LastDeployTime:     ms.LastSuccessfulDeployTime,
			BuildHistory:       ToWebViewBuildRecords(buildHistory),
			PendingBuildEdits:  pendingBuildEdits,
			PendingBuildSince:  pendingBuildSince,
			PendingBuildReason: ms.NextBuildReason(),
			CurrentBuild:       ToWebViewBuildRecord(currentBuild),
			Endpoints:          endpoints,
			PodID:              podID,
			ShowBuildStatus:    len(mt.Manifest.ImageTargets) > 0 || mt.Manifest.IsDC(),
			CombinedLog:        ms.CombinedLog,
			CrashLog:           ms.CrashLog,
			TriggerMode:        mt.Manifest.TriggerMode,
			HasPendingChanges:  hasPendingChanges,
		}

		populateResourceInfoView(mt, &r)
		r.RuntimeStatus = runtimeStatus(r.ResourceInfo())

		ret.Resources = append(ret.Resources, r)
	}

	ret.Log = s.Log
	ret.NeedsAnalyticsNudge = NeedsNudge(s)
	ret.RunningTiltBuild = TiltBuild{
		Version:   s.TiltBuildInfo.Version,
		CommitSHA: s.TiltBuildInfo.CommitSHA,
		Dev:       s.TiltBuildInfo.Dev,
		Date:      s.TiltBuildInfo.Date,
	}
	ret.LatestTiltBuild = TiltBuild{
		Version:   s.LatestTiltBuild.Version,
		CommitSHA: s.LatestTiltBuild.CommitSHA,
		Dev:       s.LatestTiltBuild.Dev,
		Date:      s.LatestTiltBuild.Date,
	}
	ret.FeatureFlags = make(map[string]bool)
	for k, v := range s.Features {
		ret.FeatureFlags[k] = v
	}
	ret.TiltCloudUsername = s.TiltCloudUsername
	ret.TiltCloudSchemeHost = cloudurl.URL(s.CloudAddress).String()
	ret.TiltCloudTeamID = s.TeamName
	if s.FatalError != nil {
		ret.FatalError = s.FatalError.Error()
	}

	return ret
}

func StateToProtoView(s store.EngineState) *proto_webview.View {
	ret := &proto_webview.View{}

	ret.Resources = append(ret.Resources, tiltfileResourceProtoView(s))

	for _, name := range s.ManifestDefinitionOrder {
		mt, ok := s.ManifestTargets[name]
		if !ok {
			continue
		}

		ms := mt.State

		var absWatchDirs []string
		var absWatchPaths []string
		for _, p := range mt.Manifest.LocalPaths() {
			fi, err := os.Stat(p)
			if err == nil && !fi.IsDir() {
				absWatchPaths = append(absWatchPaths, p)
			} else {
				absWatchDirs = append(absWatchDirs, p)
			}
		}
		absWatchPaths = append(absWatchPaths, s.TiltfilePath)
		relWatchDirs := ospath.TryAsCwdChildren(absWatchDirs)
		relWatchPaths := ospath.TryAsCwdChildren(absWatchPaths)

		var pendingBuildEdits []string
		for _, status := range ms.BuildStatuses {
			for f := range status.PendingFileChanges {
				pendingBuildEdits = append(pendingBuildEdits, f)
			}
		}

		pendingBuildEdits = ospath.FileListDisplayNames(absWatchDirs, pendingBuildEdits)

		buildHistory := append([]model.BuildRecord{}, ms.BuildHistory...)
		for i, build := range buildHistory {
			build.Edits = ospath.FileListDisplayNames(absWatchDirs, build.Edits)
			buildHistory[i] = build
		}

		currentBuild := ms.CurrentBuild
		currentBuild.Edits = ospath.FileListDisplayNames(absWatchDirs, ms.CurrentBuild.Edits)

		// Sort the strings to make the outputs deterministic.
		sort.Strings(pendingBuildEdits)

		endpoints := store.ManifestTargetEndpoints(mt)

		podID := ms.MostRecentPod().PodID

		// NOTE(nick): Right now, the UX is designed to show the output exactly one
		// pod. A better UI might summarize the pods in other ways (e.g., show the
		// "most interesting" pod that's crash looping, or show logs from all pods
		// at once).
		hasPendingChanges, pendingBuildSince := ms.HasPendingChanges()
		r := &proto_webview.Resource{
			Name:               name.String(),
			DirectoriesWatched: relWatchDirs,
			PathsWatched:       relWatchPaths,
			LastDeployTime:     ms.LastSuccessfulDeployTime.String(),
			BuildHistory:       ToProtoBuildRecords(buildHistory),
			PendingBuildEdits:  pendingBuildEdits,
			PendingBuildSince:  pendingBuildSince.String(),
			PendingBuildReason: int64(ms.NextBuildReason()),
			CurrentBuild:       ToProtoBuildRecord(currentBuild),
			Endpoints:          endpoints,
			PodID:              podID.String(),
			ShowBuildStatus:    len(mt.Manifest.ImageTargets) > 0 || mt.Manifest.IsDC(),
			CombinedLog:        ms.CombinedLog.String(),
			CrashLog:           ms.CrashLog.String(),
			TriggerMode:        triggerModeToProtoTriggerMode(mt.Manifest.TriggerMode),
			HasPendingChanges:  hasPendingChanges,
		}

		riv := protoPopulateResourceInfoView(mt, r)
		r.RuntimeStatus = riv.Status()

		ret.Resources = append(ret.Resources, r)
	}

	ret.Log = s.Log.String()
	ret.NeedsAnalyticsNudge = NeedsNudge(s)
	ret.RunningTiltBuild = &proto_webview.TiltBuild{
		Version:   s.TiltBuildInfo.Version,
		CommitSHA: s.TiltBuildInfo.CommitSHA,
		Dev:       s.TiltBuildInfo.Dev,
		Date:      s.TiltBuildInfo.Date,
	}
	ret.LatestTiltBuild = &proto_webview.TiltBuild{
		Version:   s.LatestTiltBuild.Version,
		CommitSHA: s.LatestTiltBuild.CommitSHA,
		Dev:       s.LatestTiltBuild.Dev,
		Date:      s.LatestTiltBuild.Date,
	}
	ret.FeatureFlags = make(map[string]bool)
	for k, v := range s.Features {
		ret.FeatureFlags[k] = v
	}
	ret.TiltCloudUsername = s.TiltCloudUsername
	ret.TiltCloudSchemeHost = cloudurl.URL(s.CloudAddress).String()
	ret.TiltCloudTeamID = s.TeamName
	if s.FatalError != nil {
		ret.FatalError = s.FatalError.Error()
	}

	return ret
}

func triggerModeToProtoTriggerMode(tm model.TriggerMode) proto_webview.TriggerMode {
	if tm == model.TriggerModeAuto {
		return proto_webview.TriggerMode_TRIGGER_MODE_AUTO
	}

	return proto_webview.TriggerMode_TRIGGER_MODE_MANUAL
}

func tiltfileResourceView(s store.EngineState) Resource {
	ltfb := s.TiltfileState.LastBuild()
	ctfb := s.TiltfileState.CurrentBuild
	if !ctfb.Empty() {
		ltfb.Log = ctfb.Log
	}

	ltfb.Edits = ospath.FileListDisplayNames([]string{filepath.Dir(s.TiltfilePath)}, ltfb.Edits)

	tr := Resource{
		Name:         store.TiltfileManifestName,
		IsTiltfile:   true,
		CurrentBuild: ToWebViewBuildRecord(ctfb),
		BuildHistory: []BuildRecord{
			ToWebViewBuildRecord(ltfb),
		},
		CombinedLog:   s.TiltfileState.CombinedLog,
		RuntimeStatus: RuntimeStatusOK,
	}
	if !ctfb.Empty() {
		tr.PendingBuildSince = ctfb.StartTime
	} else {
		tr.LastDeployTime = ltfb.FinishTime
	}
	return tr
}

func tiltfileResourceProto(s store.EngineState) *proto_webview.Resource {
	ltfb := s.TiltfileState.LastBuild()
	ctfb := s.TiltfileState.CurrentBuild
	if !ctfb.Empty() {
		ltfb.Log = ctfb.Log
	}

	ltfb.Edits = ospath.FileListDisplayNames([]string{filepath.Dir(s.TiltfilePath)}, ltfb.Edits)

	tr := &proto_webview.Resource{
		Name:         store.TiltfileManifestName.String(),
		IsTiltfile:   true,
		CurrentBuild: ToProtoBuildRecord(ctfb),
		BuildHistory: []*proto_webview.BuildRecord{
			ToProtoBuildRecord(ltfb),
		},
		CombinedLog:   s.TiltfileState.CombinedLog.String(),
		RuntimeStatus: string(RuntimeStatusOK),
	}
	if !ctfb.Empty() {
		tr.PendingBuildSince = ctfb.StartTime.String()
	} else {
		tr.LastDeployTime = ltfb.FinishTime.String()
	}
	return tr
}

func tiltfileResourceProtoView(s store.EngineState) *proto_webview.Resource {
	ltfb := s.TiltfileState.LastBuild()
	ctfb := s.TiltfileState.CurrentBuild
	if !ctfb.Empty() {
		ltfb.Log = ctfb.Log
	}

	ltfb.Edits = ospath.FileListDisplayNames([]string{filepath.Dir(s.TiltfilePath)}, ltfb.Edits)

	tr := &proto_webview.Resource{
		Name:         string(store.TiltfileManifestName),
		IsTiltfile:   true,
		CurrentBuild: ToProtoBuildRecord(ctfb),
		BuildHistory: []*proto_webview.BuildRecord{
			ToProtoBuildRecord(ltfb),
		},
		CombinedLog:   s.TiltfileState.CombinedLog.String(),
		RuntimeStatus: string(RuntimeStatusOK),
	}
	if !ctfb.Empty() {
		tr.PendingBuildSince = ctfb.StartTime.String()
	} else {
		tr.LastDeployTime = ltfb.FinishTime.String()
	}
	return tr
}

func populateResourceInfoView(mt *store.ManifestTarget, r *Resource) {
	if mt.Manifest.IsUnresourcedYAMLManifest() {
		r.YAMLResourceInfo = &YAMLResourceInfo{
			K8sResources: mt.Manifest.K8sTarget().DisplayNames,
		}
		return
	}

	if mt.Manifest.IsDC() {
		dc := mt.Manifest.DockerComposeTarget()
		dcState := mt.State.DCRuntimeState()
		info := NewDCResourceInfo(dc.ConfigPaths, dcState.Status, dcState.ContainerID, dcState.Log(), dcState.StartTime)
		r.DCResourceInfo = &info
		return
	}
	if mt.Manifest.IsLocal() {
		r.LocalResourceInfo = &LocalResourceInfo{}
		return
	}
	if mt.Manifest.IsK8s() {
		kState := mt.State.K8sRuntimeState()
		pod := kState.MostRecentPod()
		r.K8sResourceInfo = &K8sResourceInfo{
			PodName:            pod.PodID.String(),
			PodCreationTime:    pod.StartedAt,
			PodUpdateStartTime: pod.UpdateStartTime,
			PodStatus:          pod.Status,
			PodStatusMessage:   strings.Join(pod.StatusMessages, "\n"),
			AllContainersReady: pod.AllContainersReady(),
			PodRestarts:        pod.VisibleContainerRestarts(),
			PodLog:             pod.Log(),
		}
		return
	}

	panic("Unrecognized manifest type (not one of: k8s, DC, local)")
}

func protoPopulateResourceInfoView(mt *store.ManifestTarget, r *proto_webview.Resource) ResourceInfoView {
	if mt.Manifest.IsUnresourcedYAMLManifest() {
		r.YamlResourceInfo = &proto_webview.YAMLResourceInfo{
			K8SResources: mt.Manifest.K8sTarget().DisplayNames,
		}
		riv := &YAMLResourceInfo{
			K8sResources: mt.Manifest.K8sTarget().DisplayNames,
		}
		return riv
	}

	if mt.Manifest.IsDC() {
		dc := mt.Manifest.DockerComposeTarget()
		dcState := mt.State.DCRuntimeState()
		info := NewProtoDCResourceInfo(dc.ConfigPaths, dcState.Status, dcState.ContainerID, dcState.Log(), dcState.StartTime)
		riv := NewDCResourceInfo(dc.ConfigPaths, dcState.Status, dcState.ContainerID, dcState.Log(), dcState.StartTime)
		r.DcResourceInfo = info
		return riv
	}
	if mt.Manifest.IsLocal() {
		r.LocalResourceInfo = &proto_webview.LocalResourceInfo{}
		return &LocalResourceInfo{}
	}
	if mt.Manifest.IsK8s() {
		kState := mt.State.K8sRuntimeState()
		pod := kState.MostRecentPod()
		r.K8SResourceInfo = &proto_webview.K8SResourceInfo{
			PodName:            pod.PodID.String(),
			PodCreationTime:    pod.StartedAt.String(),
			PodUpdateStartTime: pod.UpdateStartTime.String(),
			PodStatus:          pod.Status,
			PodStatusMessage:   strings.Join(pod.StatusMessages, "\n"),
			AllContainersReady: pod.AllContainersReady(),
			PodRestarts:        int64(pod.VisibleContainerRestarts()),
			PodLog:             pod.Log().String(),
		}
		return &K8sResourceInfo{
			PodName:            pod.PodID.String(),
			PodCreationTime:    pod.StartedAt,
			PodUpdateStartTime: pod.UpdateStartTime,
			PodStatus:          pod.Status,
			PodStatusMessage:   strings.Join(pod.StatusMessages, "\n"),
			AllContainersReady: pod.AllContainersReady(),
			PodRestarts:        pod.VisibleContainerRestarts(),
			PodLog:             pod.Log(),
		}
	}

	panic("Unrecognized manifest type (not one of: k8s, DC, local)")
}

func runtimeStatus(res ResourceInfoView) RuntimeStatus {
	_, isLocal := res.(LocalResourceInfo)
	if isLocal {
		return RuntimeStatusOK
	}
	// if we have no images to build, we have no runtime status monitoring.
	_, isYAML := res.(YAMLResourceInfo)
	if isYAML {
		return RuntimeStatusOK
	}

	result, ok := runtimeStatusMap[res.Status()]
	if !ok {
		return RuntimeStatusError
	}

	return result
}

var runtimeStatusMap = map[string]RuntimeStatus{
	"Running":                          RuntimeStatusOK,
	"ContainerCreating":                RuntimeStatusPending,
	"Pending":                          RuntimeStatusPending,
	"PodInitializing":                  RuntimeStatusPending,
	"Error":                            RuntimeStatusError,
	"CrashLoopBackOff":                 RuntimeStatusError,
	"ErrImagePull":                     RuntimeStatusError,
	"ImagePullBackOff":                 RuntimeStatusError,
	"RunContainerError":                RuntimeStatusError,
	"StartError":                       RuntimeStatusError,
	string(dockercompose.StatusInProg): RuntimeStatusPending,
	string(dockercompose.StatusUp):     RuntimeStatusOK,
	string(dockercompose.StatusDown):   RuntimeStatusError,
	"Completed":                        RuntimeStatusOK,

	// If the runtime status hasn't shown up yet, we assume it's pending.
	"": RuntimeStatusPending,
}
