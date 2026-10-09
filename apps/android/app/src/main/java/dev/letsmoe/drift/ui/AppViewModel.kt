package dev.letsmoe.drift.ui

import androidx.compose.runtime.Composable
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import dev.letsmoe.drift.AppContainer

/**
 * A ViewModel scoped to the current navigation entry, built from the app's dependencies.
 * Use [key] when one entry can show different content (it is not needed for route arguments,
 * which already create a new entry).
 */
@Composable
inline fun <reified ViewModelType : ViewModel> appViewModel(
    key: String? = null,
    crossinline create: (AppContainer) -> ViewModelType,
): ViewModelType {
    val container = LocalAppContainer.current
    return viewModel(
        key = key,
        factory = viewModelFactory { initializer { create(container) } },
    )
}
