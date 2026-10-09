# gomobile bindings are reached through JNI and reflection.
-keep class dev.letsmoe.drift.mobile.** { *; }
-keep class go.** { *; }
-dontwarn go.**

# kotlinx.serialization models.
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.**
-keepclassmembers class dev.letsmoe.drift.**$$serializer { *; }
-keepclassmembers class dev.letsmoe.drift.** {
    *** Companion;
}
-keepclasseswithmembers class dev.letsmoe.drift.** {
    kotlinx.serialization.KSerializer serializer(...);
}
