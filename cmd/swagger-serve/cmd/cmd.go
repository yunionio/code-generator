package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/loads/fmts"
	"github.com/pkg/errors"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"yunion.io/x/log"
)

func NewRootCmd() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "swagger-serve",
		Short: "swagger serve for onecloud project",
	}
	cmds.AddCommand(newGenerateCmd())
	return cmds
}

type generateOption struct {
	SpecFiles []string
	CDNPrefix string
	UIVersion string
	OutputDir string
	Serve     bool
	NoOpen    bool
	ServeAddr string
	ServePort int
}

func urlJoin(prefix string, suffix ...string) string {
	u, err := url.Parse(prefix)
	if err != nil {
		log.Fatalf("Invalid url: %s", prefix)
	}
	parts := []string{u.Path}
	parts = append(parts, suffix...)
	u.Path = path.Join(parts...)
	return u.String()
}

func (o generateOption) urlJoin(filePath string) string {
	return urlJoin(o.CDNPrefix, o.UIVersion, filePath)
}

func (o generateOption) StandalonePresetJS() string {
	return o.urlJoin("swagger-ui-standalone-preset.js")
}

func (o generateOption) BundleJS() string {
	return o.urlJoin("swagger-ui-bundle.js")
}

func (o generateOption) UICss() string {
	return o.urlJoin("swagger-ui.css")
}

func (o generateOption) newUIIndexHTMLConfig() (*UIIndexHTMLConfig, error) {
	config := &UIIndexHTMLConfig{
		UICss:              o.UICss(),
		BundleJS:           o.BundleJS(),
		StandalonePresetJS: o.StandalonePresetJS(),
		URLs:               make([]*SwaggerFile, 0),
	}
	loads.AddLoader(fmts.YAMLMatcher, fmts.YAMLDoc)
	for _, spec := range o.SpecFiles {
		u, err := newSwaggerFile(spec)
		if err != nil {
			return nil, err
		}
		config.URLs = append(config.URLs, u)
	}
	return config, nil
}

func checkErr(err error) {
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
}

func newGenerateCmd() *cobra.Command {
	cfg := new(generateOption)
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "generate swagger static web site",
		Run: func(_ *cobra.Command, _ []string) {
			err := doGenerate(cfg)
			checkErr(err)
			checkErr(serveHTTP(cfg))
		},
	}
	initGenerateCmdOpts(cmd.PersistentFlags(), cfg)
	return cmd
}

func initGenerateCmdOpts(flagSet *flag.FlagSet, cfg *generateOption) {
	flagSet.StringSliceVarP(&cfg.SpecFiles, "input", "i", nil, "input swagger spec yaml or json file")
	flagSet.StringVarP(&cfg.OutputDir, "output", "o", "./_output/swagger_site", "generated swagger UI site")
	flagSet.StringVar(&cfg.CDNPrefix, "cdn", "https://cdnjs.cloudflare.com/ajax/libs/swagger-ui", "swagger-ui cdn prefix")
	flagSet.StringVar(&cfg.UIVersion, "ui-version", "3.23.11", "swagger ui version")
	flagSet.BoolVarP(&cfg.Serve, "serve", "s", false, "serve as http static server and open browser view site")
	flagSet.BoolVar(&cfg.NoOpen, "no-open", false, "Not open UI in browser")
	flagSet.StringVar(&cfg.ServeAddr, "serve-addr", "", "server listen address")
	flagSet.IntVarP(&cfg.ServePort, "serve-port", "p", 0, "server listen port, random defaultly")
}

func newSwaggerFile(specPath string) (*SwaggerFile, error) {
	spec, err := loads.Spec(specPath)
	if err != nil {
		return nil, errors.Wrapf(err, "load swagger spec %s", specPath)
	}
	info := spec.Spec().Info
	name := info.Title
	if name == "" {
		name = info.Description
	}
	return &SwaggerFile{
		Name: name,
		Path: fmt.Sprintf("%s", filepath.Base(specPath)),
	}, nil
}

func cp(srcFile, dstFile string) error {
	input, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dstFile, input, 0644); err != nil {
		return err
	}
	return nil
}

func doGenerate(cfg *generateOption) error {
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}
	for i := range cfg.SpecFiles {
		srcPath := cfg.SpecFiles[i]
		dstPath := filepath.Join(cfg.OutputDir, filepath.Base(srcPath))
		if err := cp(srcPath, dstPath); err != nil {
			return errors.Wrapf(err, "copy %s to %s", srcPath, dstPath)
		}
		cfg.SpecFiles[i] = dstPath
	}
	templateCfg, err := cfg.newUIIndexHTMLConfig()
	if err != nil {
		return err
	}
	index, err := templateCfg.Generate()
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(cfg.OutputDir, "index.html"), index, 0644); err != nil {
		return err
	}
	log.Infof("generate swagger ui site to %q\n", cfg.OutputDir)
	return nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.Wrap(err, "dial 8.8.8.8:80")
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func serveHTTP(cfg *generateOption) error {
	if !cfg.Serve {
		return nil
	}
	fs := http.FileServer(http.Dir(cfg.OutputDir))
	http.Handle("/", fs)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.ServeAddr, cfg.ServePort))
	if err != nil {
		return err
	}
	ch := make(chan bool, 0)
	go func() {
		http.Serve(listener, nil)
		ch <- true
	}()
	url := "http://" + listener.Addr().String()
	if cfg.ServeAddr == "" {
		addr, err := getLocalIP()
		if err == nil {
			url = fmt.Sprintf("http://%s:%d", addr, cfg.ServePort)
		} else {
			log.Errorf("Get local ip: %v", err)
		}
	}
	log.Infof("Serve at address: %s", url)
	for i := 0; i < 3; i++ {
		_, err = http.Get(url)
	}
	if err != nil {
		return err
	}
	if !cfg.NoOpen {
		if err := open.Run(url); err != nil {
			log.Errorf("open %s: %v", url, err)
		}
	}
	<-ch
	return nil
}
