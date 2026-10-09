# gomobile bindings are reached through JNI and reflection.
-keep class dev.neoworks.ottermusic.mobile.** { *; }
-keep class go.** { *; }
-dontwarn go.**

# kotlinx.serialization models.
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.**
-keepclassmembers class dev.neoworks.ottermusic.**$$serializer { *; }
-keepclassmembers class dev.neoworks.ottermusic.** {
    *** Companion;
}
-keepclasseswithmembers class dev.neoworks.ottermusic.** {
    kotlinx.serialization.KSerializer serializer(...);
}
