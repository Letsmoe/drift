package dev.letsmoe.drift.ui.screens.likedsongs

import dev.letsmoe.drift.core.DriftCore
import dev.letsmoe.drift.core.model.Track
import dev.letsmoe.drift.ui.screens.common.RemoteViewModel

class LikedSongsViewModel(core: DriftCore) : RemoteViewModel<List<Track>>({ core.likedTracks() })
