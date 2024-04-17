VERSION=UNKNOWN
ASSET=groovy-v$(VERSION).zip

DOWNLOADURL='https://groovy.jfrog.io/ui/api/v1/download?repoKey=dist-release-local&path=groovy-zips%252Fapache-groovy-sdk-$(VERSION).zip&isNativeBrowsing=true'

DOWNLOADZIP=apache-groovy-sdk-$(VERSION).zip
EXTRACTEDDIR=groovy-$(VERSION)

$(ASSET): $(DOWNLOADZIP)
	if test "$(VERSION)" = "UNKNOWN"; then echo -e "\n\nmust pass a version. e.g.: make VERSION=4.0.4\n\n"; false; fi
	rm -fr $(EXTRACTEDDIR)
	unzip $(DOWNLOADZIP)
	cp .bz $(EXTRACTEDDIR)
	cp .bz.lock $(EXTRACTEDDIR)
	cd $(EXTRACTEDDIR) && zip -r ../$(ASSET) .

release: $(ASSET)
	gh release delete -y v$(VERSION)
	gh release create --generate-notes -t v$(VERSION) v$(VERSION) $(ASSET)

$(DOWNLOADZIP):
	wget -O $(DOWNLOADZIP) $(DOWNLOADURL)

.PHONY: clean
clean:
	rm -fr $(DOWNLOADZIP)
	rm -fr $(ASSET)
	rm -fr $(EXTRACTEDDIR)

