package config

import (
	"encoding/json"
	"os"

	"github.com/sirupsen/logrus"
)

var _Conf Config

func SetConfig(c Config) {
	_Conf = c
}

func GetConfig() *Config {
	return &_Conf
}

type Config struct {
	DeploymentEnv          string                `json:"deployment_env"`
	ServiceName            string                `json:"service_name"`
	ServiceGroupName       string                `json:"service_group_name"`
	ServiceGrpcEndpoint    string                `json:"service_grpc_endpoint"`
	ServiceHttpEndpoint    string                `json:"service_http_endpoint"`
	ServiceMetricsEndpoint string                `json:"service_metrics_endpoint"`
	LogLevel               string                `json:"log_level"`
	LogSentryDSN           string                `json:"log_sentry_dsn"`
	LogPrinter             string                `json:"log_printer"`
	LogPrinterFilePath     string                `json:"log_printer_filepath"`
	EnableReflection       bool                  `json:"enable_reflection"`
	ServiceInternalConfig  ServiceInternalConfig `json:"service_internal_config"`
}

type SupportedApp struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

type Storage struct {
	Endpoint    string `json:"endpoint"`
	RootUsr     string `json:"root_usr"`
	RootPwd     string `json:"root_pwd"`
	EnableSSL   bool   `json:"enable_ssl"`
	DB          string `json:"db"`
	ConnTimeout int    `json:"conn_timeout"`
}

type Cache struct {
	Endpoint    string `json:"endpoint"`
	Pwd         string `json:"pwd"`
	EnableSSL   bool   `json:"enable_ssl"`
	DB          int    `json:"db"`
	ConnTimeout int    `json:"conn_timeout"`
}

type ServiceInternalConfig struct {
	MerchantID                  string         `json:"merchant_id"`
	MerchantCertSerialNo        string         `json:"merchant_cert_serial_no"`
	MerchantAPIv3Key            string         `json:"merchant_api_v3_key"`
	MerchantAPIv3SecretCertPath string         `json:"merchant_api_v3_secret_cert_path"`
	SupportedAppList            []SupportedApp `json:"supported_app_list"`
	PaymentCallbackNotifyURL    string         `json:"payment_callback_notify_url"`
	PaymentExpireTimeInMinute   int64          `json:"payment_expire_time_in_minute"`
	Storage                     Storage        `json:"storage"`
	Cache                       Cache          `json:"cache"`
}

func (conf *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &struct {
		*Alias
	}{Alias: (*Alias)(conf)}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if aux.ServiceInternalConfig.MerchantID == "MERCHANT_ID" {
		conf.ServiceInternalConfig.MerchantID = os.Getenv("MERCHANT_ID")
	}
	if aux.ServiceInternalConfig.MerchantCertSerialNo == "MERCHANT_CERT_SERIAL_NO" {
		conf.ServiceInternalConfig.MerchantCertSerialNo = os.Getenv("MERCHANT_CERT_SERIAL_NO")
	}
	if aux.ServiceInternalConfig.MerchantAPIv3Key == "MERCHANT_API_V3_KEY" {
		conf.ServiceInternalConfig.MerchantAPIv3Key = os.Getenv("MERCHANT_API_V3_KEY")
	}
	if aux.ServiceInternalConfig.MerchantAPIv3SecretCertPath == "MERCHANT_API_V3_SECRET_CERT_PATH" {
		conf.ServiceInternalConfig.MerchantAPIv3SecretCertPath = os.Getenv("MERCHANT_API_V3_SECRET_CERT_PATH")
	}
	if aux.ServiceInternalConfig.PaymentCallbackNotifyURL == "PAYMENT_CALLBACK_NOTIFY_URL" {
		conf.ServiceInternalConfig.PaymentCallbackNotifyURL = os.Getenv("PAYMENT_CALLBACK_NOTIFY_URL")
	}
	if aux.ServiceInternalConfig.SupportedAppList[0].AppID == "WX_APP_ID" {
		conf.ServiceInternalConfig.SupportedAppList[0].AppID = os.Getenv("WX_APP_ID")
	}
	if aux.ServiceInternalConfig.SupportedAppList[0].AppSecret == "WX_APP_SECRET" {
		conf.ServiceInternalConfig.SupportedAppList[0].AppSecret = os.Getenv("WX_APP_SECRET")
	}
	if aux.ServiceInternalConfig.Storage.RootPwd == "STORAGE_PWD" {
		conf.ServiceInternalConfig.Storage.RootPwd = os.Getenv("STORAGE_PWD")
	}
	if aux.ServiceInternalConfig.Cache.Pwd == "CACHE_PWD" {
		conf.ServiceInternalConfig.Cache.Pwd = os.Getenv("CACHE_PWD")
	}

	return nil
}

func loadConfigFile(fn string) error {
	data, err := os.ReadFile(fn)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &_Conf)
}

func LoadConfigFileOrPanic(fn string) *Config {
	if err := loadConfigFile(fn); err != nil {
		logrus.WithError(err).Fatalf("Failed to load config file:%s.", fn)
	} else {
		logrus.Debugf("Loaded config file:%s.", fn)
	}
	// print.PrettyPrintStruct(_Conf, 1, 4)
	return &_Conf
}
