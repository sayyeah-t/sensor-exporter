APPNAME=sensor-exporter
CONFDIR=/etc/
BINDIR=/usr/local/bin
SERVICEFILEDIR=/usr/lib/systemd/system
BASECONFDIR=etc/
DEB_WORKDIR=package

all: dep build
build:
	cd src && go build
clean:
	go clean
	rm -f src/$(APPNAME)
	rm -rf $(APPNAME)*.deb
	rm -rf $(DEB_WORKDIR)
install:
	sudo mkdir -p $(CONFDIR)
	if [ ! -e /sample.conf ]; then\
		sudo $(APPNAME).conf.sample $(CONFDIR)/$(APPNAME).conf;\
		sudo $(APPNAME).service $(SERVICEFILEDIR)/$(APPNAME).service;\
		sudo cp src/$(APPNAME) $(BINDIR);\
	fi
	sudo systemctl daemon-reload
uninstall:
	sudo systemctl stop $(APPNAME)
	sudo rm -f $(BINDIR)/$(APPNAME)
	sudo rm -rf $(CONFDIR)
	sudo rm -f $(SERVICEFILEDIR)/$(APPNAME).service
	sudo systemctl daemon-reload
dep:
	cd src && go mod tidy
package:
	mkdir -p $(DEB_WORKDIR)$(BINDIR)
	cp src/$(APPNAME) $(DEB_WORKDIR)$(BINDIR)
	mkdir -p $(DEB_WORKDIR)$(CONFDIR)
	cp $(APPNAME).conf.sample $(DEB_WORKDIR)$(CONFDIR)/$(APPNAME).conf
	mkdir -p $(DEB_WORKDIR)$(SERVICEFILEDIR)
	cp $(APPNAME).service $(DEB_WORKDIR)$(SERVICEFILEDIR)
	cp -R DEBIAN $(DEB_WORKDIR)
	fakeroot dpkg-deb --build $(DEB_WORKDIR) .
