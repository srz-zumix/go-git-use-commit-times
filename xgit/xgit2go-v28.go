// +build v28

/*
Copyright © 2020 srz_zumix <https://github.com/srz-zumix>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package xgit

import (
	git "github.com/libgit2/git2go/v28"
)

type Blob = git.Blob
type Commit = git.Commit
type Diff = git.Diff
type DiffDelta = git.DiffDelta
type DiffForEachHunkCallback = git.DiffForEachHunkCallback
type DiffOptions = git.DiffOptions
type ErrorCode = git.ErrorCode
type Odb = git.Odb
type Oid = git.Oid
type Object = git.Object
type ObjectType = git.ObjectType
type Reference = git.Reference
type Repository = git.Repository
type RevWalk = git.RevWalk
type Tree = git.Tree
type TreeEntry = git.TreeEntry

const DeltaUnmodified = git.DeltaUnmodified
const DeltaAdded = git.DeltaAdded
const DeltaDeleted = git.DeltaDeleted
const DeltaModified = git.DeltaModified
const DeltaRenamed = git.DeltaRenamed
const DeltaCopied = git.DeltaCopied
const DeltaIgnored = git.DeltaIgnored
const DeltaUntracked = git.DeltaUntracked
const DeltaTypeChange = git.DeltaTypeChange
const DeltaUnreadable = git.DeltaUnreadable
const DeltaConflicted = git.DeltaConflicted

const DiffDetailFiles = git.DiffDetailFiles
const DiffDetailHunks = git.DiffDetailHunks
const DiffDetailLines = git.DiffDetailLines

const DiffFormatPatch = git.DiffFormatPatch
const DiffFormatPatchHeader = git.DiffFormatPatchHeader
const DiffFormatRaw = git.DiffFormatRaw
const DiffFormatNameOnly = git.DiffFormatNameOnly
const DiffFormatNameStatus = git.DiffFormatNameStatus

const ErrIterOver = git.ErrIterOver

const ObjectBlob = git.ObjectBlob
const ObjectCommit = git.ObjectCommit
const ObjectTree = git.ObjectTree

const SortNone = git.SortNone
const SortReverse = git.SortReverse
const SortTime = git.SortTime
const SortTopological = git.SortTopological

const SubmoduleIgnoreNone = git.SubmoduleIgnoreNone
const SubmoduleIgnoreUntracked = git.SubmoduleIgnoreUntracked
const SubmoduleIgnoreDirty = git.SubmoduleIgnoreDirty
const SubmoduleIgnoreAll = git.SubmoduleIgnoreAll

func OpenRepository(path string) (*git.Repository, error) {
	return git.OpenRepository(path)
}

func DefaultDiffOptions() (git.DiffOptions, error) {
	return git.DefaultDiffOptions()
}

func IsErrorCode(err error, errCode git.ErrorCode) bool {
	return git.IsErrorCode(err, errCode)
}
