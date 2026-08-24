allprojects {
    repositories {
        google()
        mavenCentral()
    }
}

val newBuildDir: Directory =
    rootProject.layout.buildDirectory
        .dir("../../build")
        .get()
rootProject.layout.buildDirectory.value(newBuildDir)

subprojects {
    val newSubprojectBuildDir: Directory = newBuildDir.dir(project.name)
    project.layout.buildDirectory.value(newSubprojectBuildDir)
}
subprojects {
    project.evaluationDependsOn(":app")
}

// Fix: ensure Kotlin compiles before Java in all subprojects
// (required for image_picker_android which has Java sources referencing Kotlin-generated Pigeon classes)
gradle.projectsEvaluated {
    subprojects {
        tasks.findByName("compileDebugJavaWithJavac")
            ?.dependsOn(tasks.findByName("compileDebugKotlin") ?: return@subprojects)
        tasks.findByName("compileReleaseJavaWithJavac")
            ?.dependsOn(tasks.findByName("compileReleaseKotlin") ?: return@subprojects)
    }
}

tasks.register<Delete>("clean") {
    delete(rootProject.layout.buildDirectory)
}
