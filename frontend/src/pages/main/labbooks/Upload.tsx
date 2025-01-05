// import Dropzone from "shadcn-dropzone";

// const dropzone = () => {
//     return (
//         <Dropzone
//             onDrop={(acceptedFiles: File) => {
//                 // Do something with the files
//             }}
//         >
//             {(dropzone: DropzoneState) => (
//                 <>
//                     {dropzone.isDragAccept ? (
//                         <div className="text-sm font-medium">
//                             Drop your files here!
//                         </div>
//                     ) : (
//                         <div className="flex items-center flex-col gap-1.5">
//                             <div className="flex items-center flex-row gap-0.5 text-sm font-medium">
//                                 Upload files
//                             </div>
//                         </div>
//                     )}
//                     <div className="text-xs text-gray-400 font-medium">
//                         {dropzone.acceptedFiles.length} files uploaded so far.
//                     </div>
//                 </>
//             )}
//         </Dropzone>
//     );
// };

const Upload = () => {
    return (
        <div>
            <h1>Labbook Upload</h1>
        </div>
    );
};

export default Upload;
