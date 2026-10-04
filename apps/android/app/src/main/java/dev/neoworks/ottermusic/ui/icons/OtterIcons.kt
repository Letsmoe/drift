package dev.neoworks.ottermusic.ui.icons

import androidx.compose.ui.graphics.vector.ImageVector
import com.adamglin.PhosphorIcons
import com.adamglin.phosphoricons.Fill
import com.adamglin.phosphoricons.Regular
import com.adamglin.phosphoricons.fill.ArrowLeft as FillArrowleft
import com.adamglin.phosphoricons.fill.ArrowUpLeft as FillArrowupleft
import com.adamglin.phosphoricons.regular.ArrowUpLeft as RegularArrowupleft
import com.adamglin.phosphoricons.regular.ArrowLeft as RegularArrowleft
import com.adamglin.phosphoricons.fill.BookmarkSimple as FillBookmarksimple
import com.adamglin.phosphoricons.regular.BookmarkSimple as RegularBookmarksimple
import com.adamglin.phosphoricons.fill.Books as FillBooks
import com.adamglin.phosphoricons.regular.Books as RegularBooks
import com.adamglin.phosphoricons.fill.CaretDown as FillCaretdown
import com.adamglin.phosphoricons.regular.CaretDown as RegularCaretdown
import com.adamglin.phosphoricons.fill.Check as FillCheck
import com.adamglin.phosphoricons.regular.Check as RegularCheck
import com.adamglin.phosphoricons.fill.Clock as FillClock
import com.adamglin.phosphoricons.regular.Clock as RegularClock
import com.adamglin.phosphoricons.fill.Disc as FillDisc
import com.adamglin.phosphoricons.regular.Disc as RegularDisc
import com.adamglin.phosphoricons.fill.DotsSixVertical as FillDotssixvertical
import com.adamglin.phosphoricons.regular.DotsSixVertical as RegularDotssixvertical
import com.adamglin.phosphoricons.fill.DotsThreeVertical as FillDotsthreevertical
import com.adamglin.phosphoricons.regular.DotsThreeVertical as RegularDotsthreevertical
import com.adamglin.phosphoricons.fill.Download as FillDownload
import com.adamglin.phosphoricons.regular.Download as RegularDownload
import com.adamglin.phosphoricons.fill.Heart as FillHeart
import com.adamglin.phosphoricons.regular.Heart as RegularHeart
import com.adamglin.phosphoricons.fill.House as FillHouse
import com.adamglin.phosphoricons.regular.House as RegularHouse
import com.adamglin.phosphoricons.fill.Infinity as FillInfinity
import com.adamglin.phosphoricons.regular.Infinity as RegularInfinity
import com.adamglin.phosphoricons.fill.ListPlus as FillListplus
import com.adamglin.phosphoricons.regular.ListPlus as RegularListplus
import com.adamglin.phosphoricons.fill.MagnifyingGlass as FillMagnifyingglass
import com.adamglin.phosphoricons.regular.MagnifyingGlass as RegularMagnifyingglass
import com.adamglin.phosphoricons.fill.Minus as FillMinus
import com.adamglin.phosphoricons.regular.Minus as RegularMinus
import com.adamglin.phosphoricons.fill.MusicNote as FillMusicnote
import com.adamglin.phosphoricons.regular.MusicNote as RegularMusicnote
import com.adamglin.phosphoricons.fill.Pause as FillPause
import com.adamglin.phosphoricons.regular.Pause as RegularPause
import com.adamglin.phosphoricons.fill.PencilSimple as FillPencilsimple
import com.adamglin.phosphoricons.regular.PencilSimple as RegularPencilsimple
import com.adamglin.phosphoricons.fill.Play as FillPlay
import com.adamglin.phosphoricons.regular.Play as RegularPlay
import com.adamglin.phosphoricons.fill.Playlist as FillPlaylist
import com.adamglin.phosphoricons.regular.Playlist as RegularPlaylist
import com.adamglin.phosphoricons.fill.Plus as FillPlus
import com.adamglin.phosphoricons.regular.Plus as RegularPlus
import com.adamglin.phosphoricons.fill.Queue as FillQueue
import com.adamglin.phosphoricons.regular.Queue as RegularQueue
import com.adamglin.phosphoricons.fill.Radio as FillRadio
import com.adamglin.phosphoricons.regular.Radio as RegularRadio
import com.adamglin.phosphoricons.fill.Repeat as FillRepeat
import com.adamglin.phosphoricons.regular.Repeat as RegularRepeat
import com.adamglin.phosphoricons.fill.RepeatOnce as FillRepeatonce
import com.adamglin.phosphoricons.regular.RepeatOnce as RegularRepeatonce
import com.adamglin.phosphoricons.fill.Share as FillShare
import com.adamglin.phosphoricons.regular.Share as RegularShare
import com.adamglin.phosphoricons.fill.Shuffle as FillShuffle
import com.adamglin.phosphoricons.regular.Shuffle as RegularShuffle
import com.adamglin.phosphoricons.fill.SkipBack as FillSkipback
import com.adamglin.phosphoricons.regular.SkipBack as RegularSkipback
import com.adamglin.phosphoricons.fill.SkipForward as FillSkipforward
import com.adamglin.phosphoricons.regular.SkipForward as RegularSkipforward
import com.adamglin.phosphoricons.fill.Trash as FillTrash
import com.adamglin.phosphoricons.regular.Trash as RegularTrash
import com.adamglin.phosphoricons.fill.User as FillUser
import com.adamglin.phosphoricons.regular.User as RegularUser
import com.adamglin.phosphoricons.fill.Warning as FillWarning
import com.adamglin.phosphoricons.regular.Warning as RegularWarning
import com.adamglin.phosphoricons.fill.X as FillX
import com.adamglin.phosphoricons.regular.X as RegularX

/** Phosphor icons (the same family the desktop app uses): outline by default, [filled] for active states. */
class OtterIcon(val outline: ImageVector, val filled: ImageVector) {
    fun select(isActive: Boolean): ImageVector {
        if (isActive) {
            return filled
        }
        return outline
    }
}

object OtterIcons {
    val Home = OtterIcon(PhosphorIcons.Regular.RegularHouse, PhosphorIcons.Fill.FillHouse)
    val Search = OtterIcon(PhosphorIcons.Regular.RegularMagnifyingglass, PhosphorIcons.Fill.FillMagnifyingglass)
    val Library = OtterIcon(PhosphorIcons.Regular.RegularBooks, PhosphorIcons.Fill.FillBooks)
    val Play = OtterIcon(PhosphorIcons.Regular.RegularPlay, PhosphorIcons.Fill.FillPlay)
    val Pause = OtterIcon(PhosphorIcons.Regular.RegularPause, PhosphorIcons.Fill.FillPause)
    val Heart = OtterIcon(PhosphorIcons.Regular.RegularHeart, PhosphorIcons.Fill.FillHeart)
    val SkipBack = OtterIcon(PhosphorIcons.Regular.RegularSkipback, PhosphorIcons.Fill.FillSkipback)
    val SkipForward = OtterIcon(PhosphorIcons.Regular.RegularSkipforward, PhosphorIcons.Fill.FillSkipforward)
    val Shuffle = OtterIcon(PhosphorIcons.Regular.RegularShuffle, PhosphorIcons.Fill.FillShuffle)
    val Repeat = OtterIcon(PhosphorIcons.Regular.RegularRepeat, PhosphorIcons.Fill.FillRepeat)
    val RepeatOnce = OtterIcon(PhosphorIcons.Regular.RegularRepeatonce, PhosphorIcons.Fill.FillRepeatonce)
    val Infinity = OtterIcon(PhosphorIcons.Regular.RegularInfinity, PhosphorIcons.Fill.FillInfinity)
    val Queue = OtterIcon(PhosphorIcons.Regular.RegularQueue, PhosphorIcons.Fill.FillQueue)
    val More = OtterIcon(PhosphorIcons.Regular.RegularDotsthreevertical, PhosphorIcons.Fill.FillDotsthreevertical)
    val ChevronDown = OtterIcon(PhosphorIcons.Regular.RegularCaretdown, PhosphorIcons.Fill.FillCaretdown)
    val FillIntoField = OtterIcon(PhosphorIcons.Regular.RegularArrowupleft, PhosphorIcons.Fill.FillArrowupleft)
    val Back = OtterIcon(PhosphorIcons.Regular.RegularArrowleft, PhosphorIcons.Fill.FillArrowleft)
    val Playlist = OtterIcon(PhosphorIcons.Regular.RegularPlaylist, PhosphorIcons.Fill.FillPlaylist)
    val Radio = OtterIcon(PhosphorIcons.Regular.RegularRadio, PhosphorIcons.Fill.FillRadio)
    val Add = OtterIcon(PhosphorIcons.Regular.RegularPlus, PhosphorIcons.Fill.FillPlus)
    val Close = OtterIcon(PhosphorIcons.Regular.RegularX, PhosphorIcons.Fill.FillX)
    val Delete = OtterIcon(PhosphorIcons.Regular.RegularTrash, PhosphorIcons.Fill.FillTrash)
    val MusicNote = OtterIcon(PhosphorIcons.Regular.RegularMusicnote, PhosphorIcons.Fill.FillMusicnote)
    val Album = OtterIcon(PhosphorIcons.Regular.RegularDisc, PhosphorIcons.Fill.FillDisc)
    val Artist = OtterIcon(PhosphorIcons.Regular.RegularUser, PhosphorIcons.Fill.FillUser)
    val Share = OtterIcon(PhosphorIcons.Regular.RegularShare, PhosphorIcons.Fill.FillShare)
    val Check = OtterIcon(PhosphorIcons.Regular.RegularCheck, PhosphorIcons.Fill.FillCheck)
    val Warning = OtterIcon(PhosphorIcons.Regular.RegularWarning, PhosphorIcons.Fill.FillWarning)
    val DragHandle = OtterIcon(PhosphorIcons.Regular.RegularDotssixvertical, PhosphorIcons.Fill.FillDotssixvertical)
    val Clock = OtterIcon(PhosphorIcons.Regular.RegularClock, PhosphorIcons.Fill.FillClock)
    val AddToPlaylist = OtterIcon(PhosphorIcons.Regular.RegularListplus, PhosphorIcons.Fill.FillListplus)
    val Remove = OtterIcon(PhosphorIcons.Regular.RegularMinus, PhosphorIcons.Fill.FillMinus)
    val Edit = OtterIcon(PhosphorIcons.Regular.RegularPencilsimple, PhosphorIcons.Fill.FillPencilsimple)
    val Download = OtterIcon(PhosphorIcons.Regular.RegularDownload, PhosphorIcons.Fill.FillDownload)
    val Bookmark = OtterIcon(PhosphorIcons.Regular.RegularBookmarksimple, PhosphorIcons.Fill.FillBookmarksimple)
}
