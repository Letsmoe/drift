package dev.letsmoe.drift.ui.screens.artist

import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.ArtistPage
import dev.letsmoe.drift.ui.screens.common.RemoteViewModel

class ArtistViewModel(core: DriftCore, artistId: String) :
    RemoteViewModel<ArtistPage>({ core.artist(artistId) })
