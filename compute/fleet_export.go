package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	commonlegacy "github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/v65/common"
	core65 "github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

const (
	fleetSchemaVersion       = 2
	defaultFleetSnapshotPath = ".DATA/compute_fleet_state.json"
)

type FleetPayload struct {
	GeneratedAt      string          `json:"generatedAt"`
	GeneratedAtEpoch int64           `json:"generatedAtEpochMs"`
	SchemaVersion    int             `json:"schemaVersion"`
	Source           FleetSource     `json:"source"`
	Customers        []FleetCustomer `json:"customers"`
	Instances        []FleetInstance `json:"instances"`
}

type FleetSource struct {
	Type             string   `json:"type"`
	Auth             string   `json:"auth"`
	Profile          string   `json:"profile"`
	CustomerStrategy string   `json:"customerStrategy"`
	Regions          []string `json:"regions"`
}

type FleetCustomer struct {
	Name           string `json:"name"`
	LastImport     int64  `json:"lastImport"`
	InstanceCount  int    `json:"instanceCount"`
	ScheduledCount int    `json:"scheduledCount"`
	CompletedCount int    `json:"completedCount"`
	ChangedCount   int    `json:"changedCount"`
}

type FleetInstance struct {
	CustomerID             string                            `json:"customerId"`
	ID                     string                            `json:"ID"`
	DisplayName            string                            `json:"Display_Name"`
	Shape                  string                            `json:"Shape"`
	State                  string                            `json:"State"`
	AvailabilityDomain     string                            `json:"Availability_Domain"`
	FaultDomain            string                            `json:"Fault_Domain"`
	TimeCreated            string                            `json:"Time_Created"`
	CompartmentID          string                            `json:"Compartment_ID"`
	CompartmentName        string                            `json:"Compartment_Name"`
	TenantID               string                            `json:"Tenant_ID"`
	Region                 string                            `json:"Region"`
	HostID                 string                            `json:"Host_ID"`
	MaintenanceRebootDue   string                            `json:"Maintenance_Reboot_Due"`
	MaintenanceUTC         string                            `json:"Maintenance_UTC"`
	MaintenanceIST         string                            `json:"Maintenance_IST"`
	MaintenanceEDT         string                            `json:"Maintenance_EDT"`
	MaintenanceSource      string                            `json:"Maintenance_Source"`
	MaintenanceAction      string                            `json:"Maintenance_Action"`
	MaintenanceEventStatus string                            `json:"Maintenance_Event_Status"`
	MaintenanceCategory    string                            `json:"Maintenance_Category"`
	MaintenanceReason      string                            `json:"Maintenance_Reason"`
	MaintenanceEventOCID   string                            `json:"Maintenance_Event_OCID"`
	LiveMigrationPref      string                            `json:"Live_Migration_Preference"`
	LastRebootUTC          string                            `json:"Last_Reboot_UTC"`
	LastRebootIST          string                            `json:"Last_Reboot_IST"`
	LastRebootEDT          string                            `json:"Last_Reboot_EDT"`
	RebootEvidence         string                            `json:"Reboot_Evidence"`
	RebootStatus           string                            `json:"reboot_status"`
	PreviousStatus         string                            `json:"Previous_Status"`
	StatusChangeSignal     string                            `json:"Status_Change_Signal"`
	StatusChangedUTC       string                            `json:"Status_Changed_UTC"`
	UniqueKey              string                            `json:"uniqueKey"`
	FreeformTags           map[string]string                 `json:"freeformTags"`
	DefinedTags            map[string]map[string]interface{} `json:"definedTags"`
	LastSeenUTC            string                            `json:"Last_Seen_UTC"`
}

type fleetSnapshot struct {
	GeneratedAtEpochMs int64                          `json:"generatedAtEpochMs"`
	Instances          map[string]fleetSnapshotRecord `json:"instances"`
}

type fleetSnapshotRecord struct {
	RebootStatus     string `json:"rebootStatus"`
	StatusChangedUTC string `json:"statusChangedUTC"`
	LastRebootUTC    string `json:"lastRebootUTC"`
	LastRebootIST    string `json:"lastRebootIST"`
	LastRebootEDT    string `json:"lastRebootEDT"`
	RebootEvidence   string `json:"rebootEvidence"`
}

func BuildFleetPayload(provider common.ConfigurationProvider, regions []identity.RegionSubscription, tenancyID string, inventory []InstanceInventory, metadata RunMetadata) (FleetPayload, error) {
	now := time.Now().UTC()

	if metadata.CustomerStrategy == "" {
		metadata.CustomerStrategy = "tenancy"
	}
	if metadata.AuthType == "" {
		metadata.AuthType = "config"
	}
	if metadata.SnapshotPath == "" {
		metadata.SnapshotPath = defaultFleetSnapshotPath
	}
	customerName := strings.TrimSpace(metadata.TenancyName)
	if customerName == "" {
		customerName = tenancyID
	}

	maintenanceByInstance, err := listLatestMaintenanceEvents(provider, inventory)
	if err != nil {
		return FleetPayload{}, err
	}
	prev, err := loadFleetSnapshot(metadata.SnapshotPath)
	if err != nil {
		return FleetPayload{}, err
	}
	next := fleetSnapshot{
		GeneratedAtEpochMs: now.UnixMilli(),
		Instances:          make(map[string]fleetSnapshotRecord, len(inventory)),
	}

	instances := make([]FleetInstance, 0, len(inventory))
	changedCount := 0
	scheduledCount := 0
	completedCount := 0

	for _, entry := range inventory {
		inst := entry.Instance
		instanceID := safeStr(inst.Id)
		maintenanceEvent, hasMaintenanceEvent := maintenanceByInstance[instanceID]
		rebootDueTime := sdkTimeToTime(inst.TimeMaintenanceRebootDue)

		rebootStatus := normalizeRebootStatus(hasMaintenanceEvent, maintenanceEvent, rebootDueTime != nil)
		if rebootStatus == "Scheduled" {
			scheduledCount++
		}
		if rebootStatus == "Completed" {
			completedCount++
		}

		previousStatus := "Not Scheduled"
		statusChangedUTC := ""
		lastRebootUTC := ""
		lastRebootIST := "-"
		lastRebootEDT := "-"
		rebootEvidence := "-"

		uniqueKey := customerName + "_" + instanceID
		prevRec, hasPrev := prev.Instances[uniqueKey]
		if hasPrev {
			if prevRec.RebootStatus != "" {
				previousStatus = prevRec.RebootStatus
			}
			statusChangedUTC = prevRec.StatusChangedUTC
			if prevRec.LastRebootUTC != "" {
				lastRebootUTC = prevRec.LastRebootUTC
			}
			if prevRec.LastRebootIST != "" {
				lastRebootIST = prevRec.LastRebootIST
			}
			if prevRec.LastRebootEDT != "" {
				lastRebootEDT = prevRec.LastRebootEDT
			}
			if prevRec.RebootEvidence != "" {
				rebootEvidence = prevRec.RebootEvidence
			}
		}

		statusChangeSignal := "No Change"
		if previousStatus != rebootStatus {
			statusChangeSignal = previousStatus + " -> " + rebootStatus
			statusChangedUTC = formatTimestampUTC(now)
			changedCount++
		}

		if rebootStatus == "Completed" && previousStatus != "Completed" {
			rebootAt := rebootCompletionTime(hasMaintenanceEvent, maintenanceEvent, now)
			lastRebootUTC = formatTimestampUTC(rebootAt)
			lastRebootIST = formatTimestampLocal(rebootAt, "Asia/Kolkata", false)
			lastRebootEDT = formatTimestampLocal(rebootAt, "America/New_York", false)
			if hasMaintenanceEvent {
				rebootEvidence = fmt.Sprintf("Maintenance event reached %s (%s)", string(maintenanceEvent.LifecycleState), safeStr(maintenanceEvent.Id))
			} else {
				rebootEvidence = "Completed status observed"
			}
		}

		maintenanceUTC, maintenanceIST, maintenanceEDT := maintenanceTimestampFields(hasMaintenanceEvent, maintenanceEvent, rebootDueTime)
		maintenanceSource := ""
		if hasMaintenanceEvent {
			maintenanceSource = "Instance Maintenance Event"
		}

		maintenanceReason := ""
		if hasMaintenanceEvent {
			if safeStr(maintenanceEvent.Description) != "" {
				maintenanceReason = safeStr(maintenanceEvent.Description)
			} else {
				maintenanceReason = humanizeEnum(string(maintenanceEvent.MaintenanceReason))
			}
		}

		freeformTags := inst.FreeformTags
		if freeformTags == nil {
			freeformTags = map[string]string{}
		}
		definedTags := inst.DefinedTags
		if definedTags == nil {
			definedTags = map[string]map[string]interface{}{}
		}

		rec := FleetInstance{
			CustomerID:             customerName,
			ID:                     instanceID,
			DisplayName:            safeStr(inst.DisplayName),
			Shape:                  safeStr(inst.Shape),
			State:                  string(inst.LifecycleState),
			AvailabilityDomain:     safeStr(inst.AvailabilityDomain),
			FaultDomain:            safeStr(inst.FaultDomain),
			TimeCreated:            formatSDKTimeUTC(sdkTimeToTime(inst.TimeCreated)),
			CompartmentID:          safeStr(entry.Compartment.Id),
			CompartmentName:        safeStr(entry.Compartment.Name),
			TenantID:               tenancyID,
			Region:                 entry.RegionName,
			HostID:                 safeStr(inst.DedicatedVmHostId),
			MaintenanceRebootDue:   yesNo(rebootDueTime != nil),
			MaintenanceUTC:         maintenanceUTC,
			MaintenanceIST:         maintenanceIST,
			MaintenanceEDT:         maintenanceEDT,
			MaintenanceSource:      maintenanceSource,
			MaintenanceAction:      maintenanceAction(hasMaintenanceEvent, maintenanceEvent),
			MaintenanceEventStatus: rebootStatus,
			MaintenanceCategory:    maintenanceCategory(hasMaintenanceEvent, maintenanceEvent),
			MaintenanceReason:      maintenanceReason,
			MaintenanceEventOCID:   maintenanceEventID(hasMaintenanceEvent, maintenanceEvent),
			LiveMigrationPref:      liveMigrationPreference(inst, hasMaintenanceEvent, maintenanceEvent),
			LastRebootUTC:          lastRebootUTC,
			LastRebootIST:          lastRebootIST,
			LastRebootEDT:          lastRebootEDT,
			RebootEvidence:         rebootEvidence,
			RebootStatus:           rebootStatus,
			PreviousStatus:         previousStatus,
			StatusChangeSignal:     statusChangeSignal,
			StatusChangedUTC:       statusChangedUTC,
			UniqueKey:              uniqueKey,
			FreeformTags:           freeformTags,
			DefinedTags:            definedTags,
			LastSeenUTC:            formatTimestampUTC(now),
		}
		instances = append(instances, rec)

		next.Instances[uniqueKey] = fleetSnapshotRecord{
			RebootStatus:     rebootStatus,
			StatusChangedUTC: statusChangedUTC,
			LastRebootUTC:    lastRebootUTC,
			LastRebootIST:    lastRebootIST,
			LastRebootEDT:    lastRebootEDT,
			RebootEvidence:   rebootEvidence,
		}
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].UniqueKey < instances[j].UniqueKey
	})

	if err := saveFleetSnapshot(metadata.SnapshotPath, next); err != nil {
		return FleetPayload{}, err
	}

	payload := FleetPayload{
		GeneratedAt:      now.Format(time.RFC3339),
		GeneratedAtEpoch: now.UnixMilli(),
		SchemaVersion:    fleetSchemaVersion,
		Source: FleetSource{
			Type:             "oci-sdk",
			Auth:             metadata.AuthType,
			Profile:          metadata.Profile,
			CustomerStrategy: metadata.CustomerStrategy,
			Regions:          regionNames(regions),
		},
		Customers: []FleetCustomer{
			{
				Name:           customerName,
				LastImport:     now.UnixMilli(),
				InstanceCount:  len(instances),
				ScheduledCount: scheduledCount,
				CompletedCount: completedCount,
				ChangedCount:   changedCount,
			},
		},
		Instances: instances,
	}
	return payload, nil
}

func listLatestMaintenanceEvents(provider common.ConfigurationProvider, inventory []InstanceInventory) (map[string]core65.InstanceMaintenanceEventSummary, error) {
	out := make(map[string]core65.InstanceMaintenanceEventSummary)
	if len(inventory) == 0 {
		return out, nil
	}

	regionalCompartments := make(map[string]map[string]map[string]struct{})
	for _, entry := range inventory {
		region := entry.RegionName
		compartmentID := safeStr(entry.Compartment.Id)
		instanceID := safeStr(entry.Instance.Id)
		if region == "" || compartmentID == "" || instanceID == "" {
			continue
		}
		if regionalCompartments[region] == nil {
			regionalCompartments[region] = make(map[string]map[string]struct{})
		}
		if regionalCompartments[region][compartmentID] == nil {
			regionalCompartments[region][compartmentID] = make(map[string]struct{})
		}
		regionalCompartments[region][compartmentID][instanceID] = struct{}{}
	}

	client, err := core65.NewComputeClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}

	for region, compartments := range regionalCompartments {
		client.SetRegion(region)
		for compartmentID, wantedInstanceIDs := range compartments {
			req := core65.ListInstanceMaintenanceEventsRequest{
				CompartmentId: common.String(compartmentID),
				Limit:         common.Int(1000),
			}
			for {
				resp, err := client.ListInstanceMaintenanceEvents(context.Background(), req)
				if err != nil {
					return nil, fmt.Errorf("list instance maintenance events region=%s compartment=%s: %w", region, compartmentID, err)
				}
				for _, event := range resp.Items {
					instanceID := safeStr(event.InstanceId)
					if _, ok := wantedInstanceIDs[instanceID]; !ok {
						continue
					}
					prev, exists := out[instanceID]
					if !exists || isMaintenanceEventNewer(event, prev) {
						out[instanceID] = event
					}
				}
				if resp.OpcNextPage == nil {
					break
				}
				req.Page = resp.OpcNextPage
			}
		}
	}
	return out, nil
}

func isMaintenanceEventNewer(a core65.InstanceMaintenanceEventSummary, b core65.InstanceMaintenanceEventSummary) bool {
	at := maintenanceEventSortTime(a)
	bt := maintenanceEventSortTime(b)
	if at.After(bt) {
		return true
	}
	if bt.After(at) {
		return false
	}
	return safeStr(a.Id) > safeStr(b.Id)
}

func maintenanceEventSortTime(e core65.InstanceMaintenanceEventSummary) time.Time {
	if e.TimeWindowStart != nil {
		return e.TimeWindowStart.Time
	}
	if e.TimeCreated != nil {
		return e.TimeCreated.Time
	}
	return time.Time{}
}

func normalizeRebootStatus(hasEvent bool, event core65.InstanceMaintenanceEventSummary, hasRebootDue bool) string {
	if hasEvent {
		switch event.LifecycleState {
		case core65.InstanceMaintenanceEventLifecycleStateSucceeded:
			return "Completed"
		case core65.InstanceMaintenanceEventLifecycleStateScheduled, core65.InstanceMaintenanceEventLifecycleStateStarted, core65.InstanceMaintenanceEventLifecycleStateProcessing:
			return "Scheduled"
		case core65.InstanceMaintenanceEventLifecycleStateCanceled, core65.InstanceMaintenanceEventLifecycleStateFailed:
			return "Not Scheduled"
		default:
			return "Not Scheduled"
		}
	}
	if hasRebootDue {
		return "Scheduled"
	}
	return "Not Scheduled"
}

func rebootCompletionTime(hasEvent bool, event core65.InstanceMaintenanceEventSummary, fallback time.Time) time.Time {
	if hasEvent {
		if event.TimeFinished != nil {
			return event.TimeFinished.Time.UTC()
		}
		if event.TimeStarted != nil {
			return event.TimeStarted.Time.UTC()
		}
		if event.TimeWindowStart != nil {
			return event.TimeWindowStart.Time.UTC()
		}
	}
	return fallback.UTC()
}

func maintenanceTimestampFields(hasEvent bool, event core65.InstanceMaintenanceEventSummary, rebootDue *time.Time) (string, string, string) {
	var t *time.Time
	if hasEvent && event.TimeWindowStart != nil {
		tmp := event.TimeWindowStart.Time.UTC()
		t = &tmp
	} else if rebootDue != nil {
		tmp := rebootDue.UTC()
		t = &tmp
	}
	if t == nil {
		return "", "-", "-"
	}
	return formatTimestampUTC(*t), formatTimestampLocal(*t, "Asia/Kolkata", false), formatTimestampLocal(*t, "America/New_York", false)
}

func maintenanceAction(hasEvent bool, event core65.InstanceMaintenanceEventSummary) string {
	if !hasEvent {
		return ""
	}
	return humanizeEnum(string(event.InstanceAction))
}

func maintenanceCategory(hasEvent bool, event core65.InstanceMaintenanceEventSummary) string {
	if !hasEvent {
		return ""
	}
	return humanizeEnum(string(event.MaintenanceCategory))
}

func maintenanceEventID(hasEvent bool, event core65.InstanceMaintenanceEventSummary) string {
	if !hasEvent {
		return ""
	}
	return safeStr(event.Id)
}

func liveMigrationPreference(inst core.Instance, hasEvent bool, event core65.InstanceMaintenanceEventSummary) string {
	if hasEvent && event.InstanceAction == core65.InstanceMaintenanceEventInstanceActionRebootMigration {
		return "Use Live Migration If Possible"
	}
	if inst.AvailabilityConfig != nil && inst.AvailabilityConfig.RecoveryAction == core.InstanceAvailabilityConfigRecoveryActionStopInstance {
		return "Stop Instance After Maintenance"
	}
	return "Use Live Migration If Possible"
}

func formatSDKTimeUTC(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatTimestampUTC(t.UTC())
}

func formatTimestampUTC(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}

func formatTimestampLocal(t time.Time, tz string, withSeconds bool) string {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		if withSeconds {
			return t.UTC().Format("2006-01-02 15:04:05")
		}
		return t.UTC().Format("2006-01-02 15:04")
	}
	layout := "2006-01-02 15:04"
	if withSeconds {
		layout = "2006-01-02 15:04:05"
	}
	return t.In(loc).Format(layout)
}

func humanizeEnum(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return ""
	}
	parts := strings.Split(strings.ToLower(in), "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func yesNo(ok bool) string {
	if ok {
		return "Yes"
	}
	return "No"
}

func sdkTimeToTime(t *commonlegacy.SDKTime) *time.Time {
	if t == nil {
		return nil
	}
	tmp := t.Time.UTC()
	return &tmp
}

func regionNames(regions []identity.RegionSubscription) []string {
	out := make([]string, 0, len(regions))
	for _, r := range regions {
		if r.RegionName != nil && *r.RegionName != "" {
			out = append(out, *r.RegionName)
		}
	}
	sort.Strings(out)
	return out
}

func loadFleetSnapshot(path string) (fleetSnapshot, error) {
	state := fleetSnapshot{
		Instances: map[string]fleetSnapshotRecord{},
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return fleetSnapshot{}, fmt.Errorf("read fleet snapshot %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &state); err != nil {
		return fleetSnapshot{}, fmt.Errorf("parse fleet snapshot %s: %w", path, err)
	}
	if state.Instances == nil {
		state.Instances = map[string]fleetSnapshotRecord{}
	}
	return state, nil
}

func saveFleetSnapshot(path string, state fleetSnapshot) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create fleet snapshot dir %s: %w", dir, err)
	}
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal fleet snapshot: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("write fleet snapshot %s: %w", path, err)
	}
	return nil
}
