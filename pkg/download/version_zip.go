package download

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"github.com/gomods/athens/pkg/download/mode"
	"github.com/gomods/athens/pkg/errors"
	"github.com/gomods/athens/pkg/log"
)

// PathVersionZip URL.
const PathVersionZip = "/{module:.+}/@v/{version}.zip"

// ZipHandler implements GET baseURL/module/@v/version.zip
func ZipHandler(dp Protocol, lggr log.Entry, df *mode.DownloadFile) http.Handler {
	const op errors.Op = "download.ZipHandler"
	f := func(w http.ResponseWriter, r *http.Request) {
		mod, ver, err := getModuleParams(r, op)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(errors.Kind(err))
			return
		}
		zip, err := dp.Zip(r.Context(), mod, ver)
		if err != nil {
			severityLevel := errors.Expect(err, errors.KindNotFound, errors.KindRedirect)
			err = errors.E(op, err, severityLevel)
			lggr.SystemErr(err)
			if errors.Kind(err) == errors.KindRedirect {
				url, err := getRedirectURL(df.URL(mod), r.URL.Path)
				if err != nil {
					lggr.SystemErr(err)
					w.WriteHeader(errors.Kind(err))
					return
				}
				http.Redirect(w, r, url, errors.KindRedirect)
				return
			}
			w.WriteHeader(errors.Kind(err))
			return
		}
		defer zip.Close()

		lenBuf := new(bytes.Buffer)
		_, err = io.Copy(lenBuf, zip)
		if err != nil {
			lggr.SystemErr(errors.E(op, errors.M(mod), errors.V(ver), err))
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Length", strconv.Itoa(lenBuf.Len()))
		_, err = io.Copy(w, lenBuf)
		if err != nil {
			lggr.SystemErr(errors.E(op, errors.M(mod), errors.V(ver), err))
		}
	}
	return http.HandlerFunc(f)
}
