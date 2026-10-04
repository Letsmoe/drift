package dev.neoworks.ottermusic.ui.screens.artist

import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.ArtistPage
import dev.neoworks.ottermusic.ui.screens.common.RemoteViewModel

class ArtistViewModel(core: OtterCore, artistId: String) :
    RemoteViewModel<ArtistPage>({ core.artist(artistId) })
