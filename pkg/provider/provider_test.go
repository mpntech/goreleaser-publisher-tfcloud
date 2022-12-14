package provider

import "testing"

func Test_extractMetadataFromPath(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name         string
		args         args
		wantFile     string
		wantProvider string
		wantVersion  string
		wantOs       string
		wantArch     string
	}{
		{
			"linux amd64",
			args{"terraform-provider-okta_3.39.0-mpn_linux_amd64.zip"},
			"terraform-provider-okta_3.39.0-mpn_linux_amd64.zip",
			"okta",
			"3.39.0-mpn",
			"linux",
			"amd64",
		},
		{
			"linux amd64 dot",
			args{"./terraform-provider-okta_3.39.0-mpn7_linux_amd64.zip"},
			"terraform-provider-okta_3.39.0-mpn7_linux_amd64.zip",
			"okta",
			"3.39.0-mpn7",
			"linux",
			"amd64",
		},
		{
			"fullpath linux amd64",
			args{"/mnt/foo/bar baz/terraform-provider-okta_3.39.0-mpn_linux_amd64.zip"},
			"terraform-provider-okta_3.39.0-mpn_linux_amd64.zip",
			"okta",
			"3.39.0-mpn",
			"linux",
			"amd64",
		},
		{
			"fullpath win linux amd64",
			args{`C:\foo\bar baz\terraform-provider-okta_3.39.0-mpn_linux_amd64.zip`},
			"terraform-provider-okta_3.39.0-mpn_linux_amd64.zip",
			"okta",
			"3.39.0-mpn",
			"linux",
			"amd64",
		},
		{
			"fullpath sums",
			args{`C:\foo\bar baz\terraform-provider-okta_3.39.0-mpn_SHA256SUMS`},
			"terraform-provider-okta_3.39.0-mpn_SHA256SUMS",
			"okta",
			"3.39.0-mpn",
			"SHA256SUMS",
			"",
		},
		{
			"fullpath sums sig",
			args{`C:\foo\bar baz\terraform-provider-okta_3.39.0-mpn_SHA256SUMS.sig`},
			"terraform-provider-okta_3.39.0-mpn_SHA256SUMS.sig",
			"okta",
			"3.39.0-mpn",
			"SHA256SUMS",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotProvider, gotVersion, gotOs, gotArch := extractMetadataFromPath(tt.args.filePath)
			if gotFile != tt.wantFile {
				t.Errorf("extractMetadataFromPath() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if gotProvider != tt.wantProvider {
				t.Errorf("extractMetadataFromPath() gotProvider = %v, want %v", gotProvider, tt.wantProvider)
			}
			if gotVersion != tt.wantVersion {
				t.Errorf("extractMetadataFromPath() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			if gotOs != tt.wantOs {
				t.Errorf("extractMetadataFromPath() gotOs = %v, want %v", gotOs, tt.wantOs)
			}
			if gotArch != tt.wantArch {
				t.Errorf("extractMetadataFromPath() gotArch = %v, want %v", gotArch, tt.wantArch)
			}
		})
	}
}
