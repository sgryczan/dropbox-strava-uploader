package models

type DropBoxListFolderResponse struct {
	Entries []DropBoxEntry `json:"entries"`
	Cursor  string         `json:"cursor"`
	HasMore bool           `json:"has_more"`
}

type DropBoxEntry struct {
	Type           string `json:".tag"`
	Name           string `json:"name"`
	ID             string `json:"id"`
	ClientModified string `json:"client_modified"`
	ServerModified string `json:"server_modified"`
	Rev            string `json:"rev"`
	Size           uint64 `json:"size"`
	PathLower      string `json:"path_lower"`
	PathDisplay    string `json:"path_display"`
	PreviewURL     string `json:"preview_url"`
	IsDownloadable bool   `json:"is_downloadable"`
}

type DropBoxFileMetaData struct {
	Type           string `json:".tag"`
	Name           string `json:"name"`
	ID             string `json:"id"`
	ClientModified string `json:"client_modified"`
	ServerModified string `json:"server_modified"`
	Rev            string `json:"rev"`
	Size           uint64 `json:"size"`
	PathLower      string `json:"path_lower`
	PathDisplay    string `json:"path_display"`
	PreviewURL     string `json:"preview_url"`
	IsDownloadable bool   `json:"is_downloadable"`
}
