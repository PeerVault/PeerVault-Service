NAME ?= Peervault
TEMPLATE_SIZE ?= 20m

.PHONY: all
all: copy $(NAME)-raw.dmg $(NAME).dmg clean

copy:
	cp ../bin/peervault macos/Peervault.app/Contents/MacOS/

$(NAME)-raw.dmg:
	rm -rf $(NAME).dmg
	ln -s /Applications macos/Applications
	hdiutil create -fs HFSX -layout SPUD -size $(TEMPLATE_SIZE) "$(NAME)-raw.dmg" -srcfolder macos -format UDRW -volname "$(NAME)" -quiet

$(NAME).dmg:
	hdiutil convert "$(NAME)-raw.dmg" -quiet -format UDZO -imagekey zlib-level=9 -o $(NAME).dmg

.PHONY: clean
clean:
	rm macos/Applications
	rm macos/Peervault.app/Contents/MacOS/peervault
	rm -rf $(NAME)-raw.dmg
