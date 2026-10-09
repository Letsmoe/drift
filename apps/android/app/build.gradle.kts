import java.util.Properties

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.kotlin.serialization)
}

android {
    namespace = "dev.neoworks.ottermusic"
    compileSdk = 37

    defaultConfig {
        applicationId = "dev.neoworks.ottermusic"
        minSdk = 26
        targetSdk = 36
        versionCode = 1
        versionName = "0.1.0"
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            // Signed with the debug key so a release build installs over the debug one and keeps the library.
            signingConfig = signingConfigs.getByName("debug")
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_21
        targetCompatibility = JavaVersion.VERSION_21
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    packaging {
        jniLibs {
            // The Go core's shared libraries must stay page-aligned and loadable as-is.
            useLegacyPackaging = false
        }
    }
}

kotlin {
    jvmToolchain(21)
}

val repositoryRoot = rootProject.projectDir.resolve("../..").canonicalFile

fun androidSdkDirectory(): String {
    val localProperties = Properties()
    val localPropertiesFile = rootProject.file("local.properties")
    if (localPropertiesFile.exists()) {
        localPropertiesFile.inputStream().use { localProperties.load(it) }
    }
    val fromLocalProperties = localProperties.getProperty("sdk.dir")
    if (fromLocalProperties != null) {
        return fromLocalProperties
    }
    return System.getenv("ANDROID_HOME") ?: System.getenv("ANDROID_SDK_ROOT") ?: ""
}

// Android Studio may start Gradle without the user's shell PATH, so add the usual Go locations.
fun pathWithGoLocations(): String {
    val home = System.getProperty("user.home")
    val locations = mutableListOf(System.getenv("PATH") ?: "")
    locations.add("$home/go/bin")
    locations.add("/usr/local/go/bin")
    locations.add("/usr/lib/go/bin")
    val whichGo = providers.exec {
        commandLine("sh", "-c", "which go || true")
        isIgnoreExitValue = true
    }.standardOutput.asText.get().trim()
    if (whichGo.isNotEmpty()) {
        locations.add(File(whichGo).parent)
    }
    return locations.joinToString(File.pathSeparator)
}

// The script skips the build itself when the AAR is newer than the Go sources.
val buildGoCore = tasks.register<Exec>("buildGoCore") {
    group = "build"
    description = "Builds the Go core into app/libs/otter-core.aar with gomobile."
    workingDir = repositoryRoot
    commandLine(repositoryRoot.resolve("scripts/build-android-core.sh").absolutePath, androidSdkDirectory())
    environment("PATH", pathWithGoLocations())
}

tasks.named("preBuild") {
    dependsOn(buildGoCore)
}

dependencies {
    implementation(files("libs/otter-core.aar"))

    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.lifecycle.viewmodel.compose)
    implementation(libs.androidx.lifecycle.runtime.compose)
    implementation(libs.androidx.navigation.compose)

    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.foundation)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.material.icons.extended)
    implementation(libs.phosphor.icons)

    implementation(libs.androidx.media3.exoplayer)
    implementation(libs.androidx.media3.session)
    implementation(libs.androidx.media3.common)

    implementation(libs.coil.compose)
    implementation(libs.coil.network.okhttp)
    implementation(libs.androidx.palette)
    implementation(libs.reorderable)

    implementation(libs.androidx.profileinstaller)

    implementation(libs.kotlinx.serialization.json)
    implementation(libs.kotlinx.coroutines.android)
    implementation(libs.kotlinx.coroutines.guava)
}
