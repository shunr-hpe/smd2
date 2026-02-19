package v1

import (
	"context"
	"encoding/json"

	"github.com/openchami/fabrica/pkg/fabrica"
)

type ComponentEndpoint struct {
	APIVersion string                  `json:"apiVersion"`
	Kind       string                  `json:"kind"`
	Metadata   fabrica.Metadata        `json:"metadata"`
	Spec       ComponentEndpointSpec   `json:"spec" validate:"required"`
	Status     ComponentEndpointStatus `json:"status,omitempty"`
}

type ComponentEndpointSpec struct {
	Description string `json:"description,omitempty" validate:"max=200"`
	ID          string `json:"ID"`

	Type           string `json:"Type"`
	Domain         string `json:"Domain,omitempty"`
	FQDN           string `json:"FQDN,omitempty"`
	RedfishType    string `json:"RedfishType"`
	RedfishSubtype string `json:"RedfishSubtype"`
	MACAddr        string `json:"MACAddr,omitempty"`
	UUID           string `json:"UUID,omitempty"`
	OdataID        string `json:"OdataID"`
	RfEndpointID   string `json:"RedfishEndpointID"`

	Enabled               bool   `json:"Enabled"`
	RedfishEndpointFQDN   string `json:"RedfishEndpointFQDN,omitempty"`
	URL                   string `json:"RedfishURL,omitempty"`
	ComponentEndpointType string `json:"ComponentEndpointType"`

	RedfishChassisInfo *ComponentChassisInfo `json:"RedfishChassisInfo,omitempty"`

	RedfishSystemInfo  *ComponentSystemInfo  `json:"RedfishSystemInfo,omitempty"`
	RedfishManagerInfo *ComponentManagerInfo `json:"RedfishManagerInfo,omitempty"`
	RedfishPDUInfo     *ComponentPDUInfo     `json:"RedfishPDUInfo,omitempty"`
	RedfishOutletInfo  any                   `json:"RedfishOutletInfo,omitempty"`
}

type ComponentEndpointStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
	Ready   bool   `json:"ready"`
}

func (r *ComponentEndpoint) Validate(ctx context.Context) error {

	return nil
}

func (r *ComponentEndpoint) GetKind() string {
	return "ComponentEndpoint"
}

func (r *ComponentEndpoint) GetName() string {
	return r.Metadata.Name
}

func (r *ComponentEndpoint) GetUID() string {
	return r.Metadata.UID
}

func (r *ComponentEndpoint) IsHub() {}

type ResourceID struct {
	Oid string `json:"@odata.id"`
}
type ComponentChassisInfo struct {
	Name string `json:"Name,omitempty"`

	Actions *ChassisActions `json:"Actions,omitempty"`
}
type ComponentSystemInfo struct {
	Name       string                 `json:"Name,omitempty"`
	Actions    *ComputerSystemActions `json:"Actions,omitempty"`
	EthNICInfo []*EthernetNICInfo     `json:"EthernetNICInfo,omitempty"`
	PowerCtlInfo

	Controls      []*Control     `json:"Controls,omitempty"`
	SerialConsole *SerialConsole `json:"SerialConsole,omitempty"`
}
type PowerCtlInfo struct {
	PowerURL string          `json:"PowerURL,omitempty"`
	PowerCtl []*PowerControl `json:"PowerControl,omitempty"`
}
type PowerControl struct {
	ResourceID
	MemberId string `json:"MemberId,omitempty"`

	Name               string        `json:"Name,omitempty"`
	PowerCapacityWatts int           `json:"PowerCapacityWatts,omitempty"`
	PowerConsumedWatts interface{}   `json:"PowerConsumedWatts,omitempty"`
	OEM                *PwrCtlOEM    `json:"OEM,omitempty"`
	RelatedItem        []*ResourceID `json:"RelatedItem,omitempty"`
}
type PwrCtlOEM struct {
	Cray *PwrCtlOEMCray `json:"Cray,omitempty"`
	HPE  *PwrCtlOEMHPE  `json:"HPE,omitempty"`
}
type PwrCtlOEMCray struct {
	PowerIdleWatts  int           `json:"PowerIdleWatts,omitempty"`
	PowerLimit      *CrayPwrLimit `json:"PowerLimit,omitempty"`
	PowerResetWatts int           `json:"PowerResetWatts,omitempty"`
}
type CrayPwrLimit struct {
	Min int `json:"Min,omitempty"`
	Max int `json:"Max,omitempty"`
}
type PwrCtlOEMHPE struct {
	PowerLimit             CrayPwrLimit `json:"PowerLimit"`
	PowerRegulationEnabled bool         `json:"PowerRegulationEnabled"`
	Status                 string       `json:"Status"`
	Target                 string       `json:"Target"`
}
type EthernetNICInfo struct {
	RedfishId           string `json:"RedfishId"`
	Oid                 string `json:"@odata.id"`
	Description         string `json:"Description,omitempty"`
	FQDN                string `json:"FQDN,omitempty"`
	Hostname            string `json:"Hostname,omitempty"`
	InterfaceEnabled    *bool  `json:"InterfaceEnabled,omitempty"`
	MACAddress          string `json:"MACAddress"`
	PermanentMACAddress string `json:"PermanentMACAddress,omitempty"`
}
type Control struct {
	URL     string    `json:"URL"`
	Control RFControl `json:"Control"`
}
type ComponentManagerInfo struct {
	Name         string             `json:"Name,omitempty"`
	Actions      *ManagerActions    `json:"Actions,omitempty"`
	EthNICInfo   []*EthernetNICInfo `json:"EthernetNICInfo,omitempty"`
	CommandShell *CommandShell      `json:"CommandShell,omitempty"`
}
type ComponentPDUInfo struct {
	Name    string                    `json:"Name,omitempty"`
	Actions *PowerDistributionActions `json:"Actions,omitempty"`
}
type ChassisActions struct {
	ChassisReset ActionReset        `json:"#Chassis.Reset"`
	OEM          *ChassisActionsOEM `json:"Oem,omitempty"`
}
type ChassisActionsOEM struct {
	ChassisEmergencyPower *ActionReset `json:"#Chassis.EmergencyPower,omitempty"`
}
type ActionReset struct {
	AllowableValues []string `json:"ResetType@Redfish.AllowableValues"`
	RFActionInfo    string   `json:"@Redfish.ActionInfo"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}
type ManagerActionsOEM struct {
	ManagerFactoryReset *ActionFactoryReset `json:"#Manager.FactoryReset,omitempty"`
	CrayProcessSchedule *ActionNamed        `json:"#CrayProcess.Schedule,omitempty"`
}
type ActionFactoryReset struct {
	AllowableValues []string `json:"FactoryResetType@Redfish.AllowableValues"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}
type ActionNamed struct {
	AllowableValues []string `json:"Name@Redfish.AllowableValues"`
	Target          string   `json:"target"`
	Title           string   `json:"title,omitempty"`
}
type ComputerSystemActions struct {
	ComputerSystemReset ActionReset `json:"#ComputerSystem.Reset"`
}
type RFControl struct {
	ControlDelaySeconds int      `json:"ControlDelaySeconds"`
	ControlMode         string   `json:"ControlMode"`
	ControlType         string   `json:"ControlType"`
	Id                  string   `json:"Id"`
	Name                string   `json:"Name"`
	PhysicalContext     string   `json:"PhysicalContext"`
	SetPoint            int      `json:"SetPoint"`
	SetPointUnits       string   `json:"SetPointUnits"`
	SettingRangeMax     int      `json:"SettingRangeMax"`
	SettingRangeMin     int      `json:"SettingRangeMin"`
	Status              StatusRF `json:"Status"`
}
type StatusRF struct {
	Health       string `json:"Health"`
	HealthRollUp string `json:"HealthRollUp,omitempty"`
	State        string `json:"State,omitempty"`
}
type SerialConsole struct {
	MaxConcurrentSessions int                    `json:"MaxConcurrentSessions"`
	IPMI                  *SerialConsoleProtocol `json:"IPMI,omitempty"`
	SSH                   *SerialConsoleProtocol `json:"SSH,omitempty"`
	Telnet                *SerialConsoleProtocol `json:"Telnet,omitempty"`
	WebSocket             *WebSocketConsole      `json:"WebSocket,omitempty"`
}
type SerialConsoleProtocol struct {
	ServiceEnabled        bool   `json:"ServiceEnabled"`
	Port                  int    `json:"Port,omitempty"`
	HotKeySequenceDisplay string `json:"HotKeySequenceDisplay,omitempty"`
	SharedWithManagerCLI  bool   `json:"SharedWithManagerCLI,omitempty"`
	ConsoleEntryCommand   string `json:"ConsoleEntryCommand,omitempty"`
}
type WebSocketConsole struct {
	ServiceEnabled bool   `json:"ServiceEnabled"`
	Interactive    bool   `json:"Interactive"`
	ConsoleURI     string `json:"ConsoleURI"`
}
type ManagerActions struct {
	ManagerReset ActionReset        `json:"#Manager.Reset"`
	OEM          *ManagerActionsOEM `json:"Oem,omitempty"`
}
type CommandShell struct {
	ServiceEnabled        bool     `json:"ServiceEnabled"`
	MaxConcurrentSessions int      `json:"MaxConcurrentSessions"`
	ConnectTypesSupported []string `json:"ConnectTypesSupported"`
}
type PowerDistributionActions struct {
	OEM *json.RawMessage `json:"Oem,omitempty"`
}
