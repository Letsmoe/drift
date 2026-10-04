package dev.neoworks.ottermusic.ui.screens.likedsongs

import dev.neoworks.ottermusic.core.OtterCore
import dev.neoworks.ottermusic.core.model.Track
import dev.neoworks.ottermusic.ui.screens.common.RemoteViewModel

class LikedSongsViewModel(core: OtterCore) : RemoteViewModel<List<Track>>({ core.likedTracks() })
