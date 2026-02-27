.PHONY: deploy build-android lint

build-android:
	./gradlew :composeApp:assembleDebug

deploy: build-android
	adb install -r composeApp/build/outputs/apk/debug/composeApp-debug.apk
	adb shell am start -n com.sensio.notification/.MainActivity

lint:
	./gradlew ktlintCheck

format:
	./gradlew ktlintFormat
